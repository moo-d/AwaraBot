import { spawn, execSync, ChildProcess, StdioOptions } from 'child_process'
import path from 'path'
import fs from 'fs'
import os from 'os'
import * as qrcode from 'qrcode-terminal'
import { Bot, Command, CommandContext, CommandResponse } from './types'
import { extractCommand } from './utils/prefix'
import { commandCache, getCommand } from './commands'
import { _aliasRegistry, _commandRegistry } from './utils/loader'
import { createBotClient } from './core/botClient'

type SupportedPlatform = 'win32' | 'linux' | 'darwin'

class WhatsAppBotLauncher {
  private static readonly BINARY_DIR = path.join(__dirname, '../bin')
  private static readonly BUILD_DIR = path.join(__dirname, '..')

  public static async launch() {
    try {
      const { initializeLoader } = await import('./utils/loader')
      await initializeLoader()
      
      this.ensureBinaryExists()
      const botProcess = this.spawnBot(this.getPlatformBinaryPath())
      this.setupEventHandlers(botProcess)
      
      const { _commandRegistry } = await import('./utils/loader')
      console.log('Loaded commands:', [..._commandRegistry.keys()])
    } catch (error) {
      console.error('❌ Fatal error:', error)
      process.exit(1)
    }
  }

  private static getPlatformBinaryPath(): string {
    const platform = os.platform() as SupportedPlatform
    const arch = os.arch()
    
    const binaries = {
      win32: `whatsapp-bot.exe`,
      linux: `whatsapp-bot-linux-${arch === 'x64' ? 'x64' : 'arm64'}`,
      darwin: `whatsapp-bot-macos-${arch === 'arm64' ? 'arm64' : 'x64'}`
    }

    const binaryPath = path.join(this.BINARY_DIR, binaries[platform])
    console.log("DEBUG BIN", binaryPath)
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
      win32: `set GOOS=windows&& set GOARCH=amd64&& go build -o ${path.join(this.BINARY_DIR, "whatsapp-bot.exe")} main.go`,
      linux: `GOOS=linux GOARCH=${arch === 'x64' ? 'amd64' : 'arm64'} go build -o "${outputFile}" main.go`,
      darwin: `GOOS=darwin GOARCH=${arch === 'arm64' ? 'arm64' : 'amd64'} go build -o "${outputFile}" main.go`
    }

    console.log(this.BUILD_DIR)
    try {
      execSync(buildCommands[platform], {
        cwd: this.BUILD_DIR,
        stdio: 'inherit',
        env: { ...process.env }
      })
    } catch (error) {
      console.error('❌ Build failed:', error)
      throw new Error('Failed to build binary')
    }
  }

  private static spawnBot(binaryPath: string): ChildProcess {
    if (!fs.existsSync(binaryPath)) {
      throw new Error(`Binary not found at ${binaryPath}`);
    }
    
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
    let messageBuffer = ''
    const bot: Bot = createBotClient(botProcess)

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

    botProcess.stdout?.on('data', (data) => {
      messageBuffer += data.toString()
      const messages = messageBuffer.split('\n')
      messageBuffer = messages.pop() || ''
      
      messages.forEach(msg => {
        if (msg.trim()) {
          try {
            handleOutput(Buffer.from(msg))
          } catch (err) {
            console.error('IPC message processing error:', err)
          }
        }
      })
    })

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
    process.on('exit', () => {
      commandCache.clear()
      _commandRegistry.clear()
      _aliasRegistry.clear()
    })
  }

  private static async handleMessage(bot: Bot, content: any) {
    const { command: cmdName, args } = extractCommand(content.text)
    if (!cmdName) return
  
    try {
      const cmd = await getCommand(cmdName)
      if (!cmd) return
      
      const context: CommandContext = {
        chat: content.chat,
        from: content.from,
        pushName: content.pushName,
        isGroup: content.isGroup
      }

      const startTime = Date.now()
      try {
        console.log(`[MSG] From: ${context.from} - Content: ${content.text}`)
        await cmd.handler(bot, args, context)
        const duration = Date.now() - startTime
        if (duration > 1000) {
          console.log(`[PERF] Slow command ${cmdName}: ${duration}ms`)
        }
      } catch (err) {
        console.error(`[ERROR] Command ${cmdName} failed after ${Date.now() - startTime}ms`, err)
        throw err
      }
    } catch (err) {
      await bot.sendMessage(content.chat, '⚠️ An error occurred while processing your command')
    }
  }
}

WhatsAppBotLauncher.launch()
