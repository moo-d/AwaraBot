import { Command, CommandWithMeta } from '../types'
import { _aliasRegistry, _commandRegistry } from '../utils/loader'

export const commandCache = new Map<string, CommandWithMeta>()

export const getCommand = async (name: string): Promise<CommandWithMeta | null> => {
  const normalizedName = name.toLowerCase()
  
  if (commandCache.has(normalizedName)) {
    return commandCache.get(normalizedName)!
  }

  const commandName = _aliasRegistry.get(normalizedName) || normalizedName
  const meta = _commandRegistry.get(commandName)
  if (!meta) return null

  try {
    delete require.cache[require.resolve(meta.filePath)]
    const cmd = (await import(meta.filePath)).default
    
    const commandWithMeta: CommandWithMeta = {
      ...cmd,
      meta: {
        ...cmd.meta,
        filePath: meta.filePath,
        category: cmd.category || meta.category
      }
    }
    
    commandCache.set(commandName, commandWithMeta)
    return commandWithMeta
  } catch (err) {
    console.error(`[CMD] Failed to load ${commandName}:`, err)
    return null
  }
}

export const handleCommand = async (
  bot: any,
  name: string, 
  args: string[],
  context: any
) => {
  const command = await getCommand(name)
  if (!command) return null
  
  try {
    const result = await command.handler(bot, args, {
      ...context,
      category: command.meta?.category || 'general'
    })
    
    return result
  } catch (err) {
    console.error(`[CMD] Execution error (${name}):`, err)
    return null
  }
}