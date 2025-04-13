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
  sendReaction: (
    jid: string,
    sender: string,
    messageId: string,
    emoji: string
  ) => Promise<void>
  ai: (
    jid: string,
    prompt: string,
    messages?: Array<{
      role: string
      content: string
    }>,
    model?: string
  ) => Promise<AIResponse>
  downloader: (
    url: string, 
    type: 'tiktok' | 'youtube', 
    format?: 'mp3' | 'mp4'
  ) => Promise<DownloadResult>
}

export interface DownloadResult {
  status: boolean
  type?: string
  result?: {
    video?: string
    images?: string[]
    music?: string
    wm?: string
    url?: string
    title?: string
    duration?: number
    [key: string]: any
  }
  error?: string
}

export interface AIResponse {
  chat: string
  message: string
  command?: string
  query?: string
  caption?: string
  isGroup?: boolean
  messageId?: string
}

export interface CommandContext {
  chat: string
  from: string
  sender: string
  text: string
  pushName?: string
  isGroup?: boolean
  messageId: string
  isImage?: boolean
  isQuotedImage?: boolean
  quotedMessage?: QuotedMessage
}

export interface QuotedMessage {
  messageId: string
  from: string
  isImage?: boolean
  isVideo?: boolean
  isDocument?: boolean
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
  wait?: boolean
  handler: (
    bot: Bot,
    args: string[],
    context: CommandContext
  ) => Promise<CommandResponse | void> | CommandResponse | void
  meta?: Partial<CommandMeta>
}

export interface CommandWithMeta extends Command {
  meta: CommandMeta
  wait?: boolean
}