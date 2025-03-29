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

export interface Command {
  name: string
  alias?: string[]
  category: string
  description?: string
  __filePath?: string
  handler: (
    bot: Bot,
    args: string[],
    context: CommandContext
  ) => Promise<CommandResponse | void> | CommandResponse | void
  meta?: {
    filePath?: string
    loadedAt?: Date
  }
}