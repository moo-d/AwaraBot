import { loadCategorizedCommands } from '../utils/loader'

const commands = loadCategorizedCommands()

export const getCommand = (name: string) => commands.get(name.toLowerCase())

export const handleCommand = async (
  bot: any,
  name: string, 
  args: string[],
  context: any
) => {
  const command = getCommand(name)
  if (!command) return null
  
  try {
    return await command.handler(bot, args, {
      ...context,
      category: command.category
    })
  } catch (err) {
    console.error(`Error executing ${name}:`, err)
    return null
  }
}
