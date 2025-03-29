import fs from 'fs/promises'
import path from 'path'
import { Command, CommandMeta, CommandWithMeta } from '../types'

export const _commandRegistry = new Map<string, CommandMeta>()
export const _aliasRegistry = new Map<string, string>()
const _commandCache = new Map<string, CommandWithMeta>()

const MAX_CACHE_SIZE = 100
const CACHE_TTL = 3600000
const PARALLEL_LIMIT = 5

function cleanupCache() {
  const now = Date.now()
  const entries = [..._commandCache.entries()]
    .filter(([, cmd]) => cmd.meta.loadedAt && now - cmd.meta.loadedAt.getTime() > CACHE_TTL)

  if (_commandCache.size > MAX_CACHE_SIZE) {
    entries.sort((a, b) => a[1].meta.loadedAt!.getTime() - b[1].meta.loadedAt!.getTime())
    entries.slice(0, entries.length - MAX_CACHE_SIZE)
      .forEach(([key]) => _commandCache.delete(key))
  }
}

setInterval(cleanupCache, CACHE_TTL).unref()

async function processFile(file: string, commandDir: string) {
  if (!file.endsWith('.js') || !file.includes('_')) return
  
  const filePath = path.join(commandDir, file)
  try {
    const stats = await fs.stat(filePath)
    const [category] = file.split('_')
    
    delete require.cache[require.resolve(filePath)]
    const cmd = (await import(filePath)).default

    if (!cmd?.name || !cmd.handler) {
      console.warn(`⚠️ Invalid command in ${file}`)
      return
    }

    const meta: CommandMeta = {
      filePath,
      loadedAt: new Date(),
      lastModified: stats.mtimeMs,
      size: stats.size,
      category: cmd.category || category
    }

    _commandRegistry.set(cmd.name.toLowerCase(), meta)
    cmd.alias?.forEach((alias: string) => _aliasRegistry.set(alias.toLowerCase(), cmd.name.toLowerCase()))
  } catch (err) {
    console.error(`✗ Failed to load ${file}:`, err)
  }
}

export async function initializeLoader(): Promise<void> {
  const commandDir = path.resolve(__dirname, '../../dist/commands')
  
  try {
    const files = (await fs.readdir(commandDir))
      .filter(file => file.endsWith('.js') && file.includes('_'))

    for (let i = 0; i < files.length; i += PARALLEL_LIMIT) {
      await Promise.all(files.slice(i, i + PARALLEL_LIMIT)
        .map(file => processFile(file, commandDir)))
    }
  } catch (err) {
    console.error('[LOADER] Initialization error:', err)
    throw err
  }
}

export async function getCommand(name: string): Promise<CommandWithMeta | null> {
  const normalizedName = name.toLowerCase()
  const cached = _commandCache.get(normalizedName)
  if (cached) return cached

  const commandName = _aliasRegistry.get(normalizedName) || normalizedName
  const meta = _commandRegistry.get(commandName)
  if (!meta) return null

  try {
    const stats = await fs.stat(meta.filePath)
    const isModified = meta.lastModified !== stats.mtimeMs

    if (isModified) {
      delete require.cache[require.resolve(meta.filePath)]
      meta.lastModified = stats.mtimeMs
      meta.loadedAt = new Date()
    }

    const cmd = (await import(meta.filePath)).default
    const commandWithMeta: CommandWithMeta = {
      ...cmd,
      meta: { ...meta, category: cmd.category || meta.category }
    }

    _commandCache.set(commandName, commandWithMeta)
    cmd.alias?.forEach((alias: string) => _commandCache.set(alias.toLowerCase(), commandWithMeta))

    return commandWithMeta
  } catch (err) {
    console.error(`[LOADER] Failed to load ${commandName}`, err)
    _commandRegistry.delete(commandName)
    return null
  }
}
