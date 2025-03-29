import { spawn, execSync, ChildProcess, StdioOptions } from 'child_process'
import path from 'path'
import fs from 'fs'
import os from 'os'
import * as qrcode from 'qrcode-terminal'
import { Bot, Command, CommandContext, CommandResponse } from './types'
import { extractCommand } from './utils/prefix'

type SupportedPlatform = 'win32' | 'linux' | 'darwin'

class WhatsAppBotLauncher {
  private static readonly BINARY_DIR = path.join(__dirname, '../bin')
  private static readonly BUILD_DIR = path.join(__dirname, '../connection')
  private static commands = new Map<string, Command>()

  public static async launch() {
    try {
      await this.loadCommands()
      this.ensureBinaryExists()
      const botProcess = this.spawnBot(this.getPlatformBinaryPath())
      this.setupEventHandlers(botProcess)
    } catch (error) {
      console.error('❌ Fatal error:', error instanceof Error ? error.message : error)
      process.exit(1)
    }
  }

  private static async loadCommands() {
    const commandsDir = path.join(__dirname, 'commands')
    try {
      const commandFiles = fs.readdirSync(commandsDir)
        .filter(file => (file.endsWith('.js') || (process.env.NODE_ENV === 'development' && file.endsWith('.ts'))))
        .filter(file => !file.startsWith('_'))

      await Promise.all(commandFiles.map(async (file) => {
        try {
          const command = await this.loadCommandFile(path.join(commandsDir, file))
          if (command) this.registerCommand(command)
        } catch (err) {
          console.error(`Failed to load ${file}:`, err)
        }
      }))
    } catch (err) {
      console.error('Command loading failed:', err)
    }
  }

  private static async loadCommandFile(filePath: string): Promise<Command | null> {
    const imported = await import(filePath)
    const command = imported?.default || imported
    return command?.name ? command : null
  }

  private static registerCommand(command: Command) {
    this.commands.set(command.name.toLowerCase(), command)
    command.alias?.forEach(alias => this.commands.set(alias.toLowerCase(), command))
  }

  private static getPlatformBinaryPath(): string {
    const platform = os.platform() as SupportedPlatform
    const arch = os.arch()
    
    const binaries = {
      win32: 'whatsapp-bot.exe',
      linux: `whatsapp-bot-linux-${arch === 'x64' ? 'x64' : 'arm64'}`,
      darwin: `whatsapp-bot-macos-${arch === 'arm64' ? 'arm64' : 'x64'}`
    }

    const binaryPath = path.join(this.BINARY_DIR, binaries[platform])
    if (!fs.existsSync(binaryPath)) throw new Error(`Binary not found at ${binaryPath}`)
    return binaryPath
  }

  private static ensureBinaryExists() {
    const binaryPath = this.getPlatformBinaryPath()
    if (!fs.existsSync(binaryPath)) {
      console.log('🔨 Building binary...')
      this.buildBinary()
    }

    if (os.platform() !== 'win32') {
      try {
        fs.chmodSync(binaryPath, 0o755)
      } catch (err) {
        console.warn('⚠️ Could not set execute permissions:', err)
      }
    }
  }

  private static buildBinary() {
    const platform = os.platform() as SupportedPlatform
    const arch = os.arch()
    const outputFile = path.join(this.BINARY_DIR, `whatsapp-bot-${platform}-${arch === 'x64' ? 'x64' : 'arm64'}`)

    const buildCommands = {
      win32: `GOOS=windows GOARCH=amd64 go build -o "${outputFile}.exe" main.go`,
      linux: `GOOS=linux GOARCH=${arch === 'x64' ? 'amd64' : 'arm64'} go build -o "${outputFile}-linux-${arch === 'x64' ? 'x64' : 'arm64'}" main.go`,
      darwin: `GOOS=darwin GOARCH=${arch === 'arm64' ? 'arm64' : 'amd64'} go build -o "${outputFile}-macos-${arch === 'arm64' ? 'arm64' : 'x64'}" main.go`
    }

    try {
      execSync(buildCommands[platform], {
        cwd: this.BUILD_DIR,
        stdio: 'inherit',
        env: { ...process.env, CGO_ENABLED: '0' }
      })
    } catch (error) {
      console.error('❌ Build failed:', error)
      throw new Error('Failed to build binary')
    }
  }

  private static spawnBot(binaryPath: string): ChildProcess {
    const botProcess = spawn(binaryPath, [], {
      cwd: path.dirname(binaryPath),
      stdio: ['pipe', 'pipe', 'inherit'] as StdioOptions,
      windowsHide: true,
      shell: false
    })

    if (os.platform() === 'win32') execSync('chcp 65001', { stdio: 'ignore' })
    return botProcess
  }

  private static setupEventHandlers(botProcess: ChildProcess) {
    const bot: Bot = {
      sendMessage: async (jid, content) => {
        const message = typeof content === 'string' ? content : ""
        const command = `SEND:${jid}|${message.replace(/\n/g, '{{NL}}')}MESSAGE_END\n`
        return new Promise(resolve => botProcess.stdin?.write(command, err => {
          if (err) console.error('Write error:', err)
          resolve()
        }))
      }
    }

    const handleOutput = (data: Buffer) => {
      const output = data.toString().trim()
      if (!output) return

      try {
        const message = JSON.parse(output)
        switch (message.type) {
          case "qr":
            qrcode.generate(message.content.code, { small: true })
            console.log(message.content.message)
            break
          case "message":
            this.handleMessage(bot, message.content)
            break
          default:
            console.log(output)
        }
      } catch {
        console.log(output)
      }
    }

    botProcess.stdout?.on('data', handleOutput)
    botProcess.stderr?.on('data', data => console.error(`[BOT ERROR] ${data.toString().trim()}`))
    botProcess.on('error', err => {
      console.error('🔥 Process error:', err)
      process.exit(1)
    })
    botProcess.on('exit', code => {
      console.log(`🛑 Process exited with code ${code}`)
      process.exit(code || 0)
    })
    process.on('SIGINT', () => {
      console.log('\nShutting down...')
      botProcess.kill()
      process.exit()
    })
  }

  private static async handleMessage(bot: Bot, content: any) {
    const messageText = content.text?.trim() || ''
    const { command: cmdName, args } = extractCommand(messageText)
    if (!cmdName) return

    const cmd = this.findCommand(cmdName)
    if (!cmd) return

    try {
      const context: CommandContext = {
        chat: content.chat,
        from: `${content.from.split(':')[0]}`,
        pushName: content.pushName,
        isGroup: content.isGroup
      }

      const response = await cmd.handler(bot, args, context)
      const senderJid = context.from.replace(/@s\.whatsapp\.net$/, '') + '@s.whatsapp.net'
      console.log(`[MSG] From: ${senderJid} - Content: ${messageText}`)
      if (response) await bot.sendMessage(content.chat, response.text)
    } catch (err) {
      console.error('Command execution error:', err)
      await bot.sendMessage(content.chat, '⚠️ Error processing command')
    }
  }

  private static findCommand(cmdName: string): Command | undefined {
    return this.commands.get(cmdName.toLowerCase()) || 
           Array.from(this.commands.values()).find(c => 
             c.alias?.includes(cmdName.toLowerCase())) ||
           (this.commands.has('default') ? 
             this.commands.get('default') : undefined)
  }
}

WhatsAppBotLauncher.launch()
