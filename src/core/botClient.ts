import { ChildProcess } from 'child_process'
import { Bot } from '../types'

export function createBotClient(botProcess: ChildProcess): Bot {
  const sendCommand = (command: string, errorPrefix = 'Command') => {
    return new Promise<void>(resolve => {
      botProcess.stdin?.write(command, err => {
        if (err) console.error(`${errorPrefix} error:`, err)
        resolve()
      })
    })
  }

  const formatContent = (content: string) => content.replace(/\n/g, '{{NL}}')

  return {
    sendMessage: async (jid, content) => {
      const message = typeof content === 'string' ? content : ''
      const command = `SEND:${jid}|${formatContent(message)}MESSAGE_END\n`
      return sendCommand(command, 'Write')
    },
    sendImage: async (jid, image, caption = '', isUrl = false) => {
      const baseCmd = isUrl || typeof image === "string" ? 'SEND_URL_IMAGE' : 'SEND_IMAGE'
      const media = typeof image === 'string' ? image : image.toString('base64')
      const command = `${baseCmd}:${jid}|${media}|${formatContent(caption)}MESSAGE_END\n`
      return sendCommand(command, 'Image send')
    },
    sendVideo: async (jid, video, caption = '', isUrl = false) => {
      const baseCmd = isUrl || typeof video === "string" ? 'SEND_URL_VIDEO' : 'SEND_VIDEO'
      const media = typeof video === 'string' ? video : video.toString('base64')
      const command = `${baseCmd}:${jid}|${media}|${formatContent(caption)}MESSAGE_END\n`
      return sendCommand(command, 'Video send')
    },
    sendAudio: async (jid, audio, isUrl = false) => {
      const baseCmd = isUrl || typeof audio === "string" ? 'SEND_URL_AUDIO' : 'SEND_AUDIO'
      const media = typeof audio === 'string' ? audio : audio.toString('base64')
      const command = `${baseCmd}:${jid}|${media}MESSAGE_END\n`
      return sendCommand(command, 'Audio send')
    }
  }
}