import { Command } from '../types'
import { _commandRegistry } from '../utils/loader'

export default {
  name: 'menu',
  alias: ['help'],
  category: 'main',
  description: 'Show all available commands',
  async handler(bot, args, context) {
    const commandsByCategory: Record<string, string[]> = {}
    
    for (const [cmdName, meta] of _commandRegistry) {
      if (!commandsByCategory[meta.category]) {
        commandsByCategory[meta.category] = []
      }
      
      let commandEntry = `• ${cmdName}`
      if (meta.alias && meta.alias.length > 0) {
        commandEntry += ` (${meta.alias.join(', ')})`
      }
      
      commandsByCategory[meta.category].push(commandEntry)
    }

    let menuMessage = '╭━━━〔 *вoт мenυ* 〕━━━╮\n\n'
    
    for (const [category, commands] of Object.entries(commandsByCategory)) {
      menuMessage += `❏ *${category.toUpperCase()}࿐",*\n`
      menuMessage += commands.join('\n')
      menuMessage += '\n\n'
    }

    menuMessage += `╰━━━━━━━━━━━━━━━━━━━╯\n`
    menuMessage += `> _Type /help <command> for more info_`

    await bot.sendMessage(context.chat, menuMessage)
  }
} as Command
