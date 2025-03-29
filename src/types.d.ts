export interface Bot {
  sendMessage: (jid: string, message: string) => Promise<void>
}

export interface CommandContext {
  chat: string
  from: string
  pushName?: string
  isGroup?: boolean
}

export interface CommandResponse {
  text: string
  mentions?: string[]
}

export interface CommandMeta {
  alias?: string[]
  filePath: string
  loadedAt?: Date
  lastModified: number
  size: number
  category: string
}

export interface Command {
  name: string
  alias?: string[]
  category: string
  description?: string
  handler: (
    bot: Bot,
    args: string[],
    context: CommandContext
  ) => Promise<CommandResponse | void> | CommandResponse | void
  meta?: Partial<CommandMeta>
}

export interface CommandWithMeta extends Command {
  meta: CommandMeta
}
