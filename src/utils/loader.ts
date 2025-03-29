import fs from 'fs'
import path from 'path'
import { Command } from '../types'

export function loadCategorizedCommands(): Map<string, Command> {
  const commandMap = new Map<string, Command>()
  const commandDir = path.join(__dirname, '../commands')

  const files = fs.readdirSync(commandDir)
    .filter(file => 
      (file.endsWith('.js') || (process.env.NODE_ENV === 'development' && file.endsWith('.ts'))) &&
      !file.startsWith('_') && 
      !['index.js', 'index.ts', 'types.js', 'types.ts'].includes(file) &&
      file.includes('_')
    )

  for (const file of files) {
    try {
      const filePath = path.join(commandDir, file)
      const modulePath = path.resolve(filePath)
      delete require.cache[require.resolve(modulePath)]
      
      const imported = require(modulePath)
      const command = imported?.default || imported

      if (!command?.name) {
        console.error(`Invalid command in ${file}: missing 'name' property`)
        continue
      }

      const [category] = file.split('_')
      command.category = category.toLowerCase()
      commandMap.set(command.name.toLowerCase(), command)

      if (command.alias) {
        command.alias.forEach((alias: string) => {
          commandMap.set(alias.toLowerCase(), command)
        })
      }
    } catch (err) {
      console.error(`Failed to load ${file}:`, err)
    }
  }

  return commandMap
        }
