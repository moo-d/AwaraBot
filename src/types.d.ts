export interface Bot {
  sendCommand: (command: string, errorPrefix?: string) => Promise<void>
  sendMessage: (jid: string, message: string) => Promise<void>
  sendImage: (
    jid: string, 
    image: Buffer | string, 
    caption?: string, 
    isUrl?: boolean
  ) => Promise<void>
  sendVideo: (
    jid: string, 
    video: Buffer | string, 
    caption?: string, 
    isUrl?: boolean
  ) => Promise<void>
  sendAudio: (
    jid: string, 
    audio: Buffer | string, 
    isUrl?: boolean
  ) => Promise<void>
  downloader: (
    url: string, 
    type: 'tiktok' | 'youtube', 
    format?: 'mp3' | 'mp4'
  ) => Promise<{
    status: boolean
    type?: string
    result?: {
      video?: string
      music?: string
      wm?: string
      url?: string
      title?: string
      duration?: number
      [key: string]: any
    }
    error?: string
  }>
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
