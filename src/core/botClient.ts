import { ChildProcess } from 'child_process'
import { Bot } from '../types'

export function createBotClient(botProcess: ChildProcess): Bot {
  const sendCommand = (command: string, errorPrefix = 'Command') => {
    return new Promise<void>((resolve, reject) => {
      botProcess.stdin?.write(command, err => {
        err ? reject(`${errorPrefix} error: ${err}`) : resolve()
      })
    })
  }

  const formatContent = (content: string) => content.replace(/\n/g, '{{NL}}')
  const createMediaCommand = (
    type: 'IMAGE'|'VIDEO'|'AUDIO',
    jid: string,
    media: string|Buffer,
    caption = '',
    isUrl = false
  ) => {
    const baseCmd = isUrl || typeof media === 'string' ? `SEND_URL_${type}` : `SEND_${type}`
    const mediaData = typeof media === 'string' ? media : media.toString('base64')
    return `${baseCmd}:${jid}|${mediaData}${type !== 'AUDIO' ? `|${formatContent(caption)}` : ''}MESSAGE_END\n`
  }

  const handleDownloadResponse = (): Promise<any> => {
    return new Promise((resolve, reject) => {
      const handler = (data: Buffer) => {
        const message = data.toString()
        if (message.includes('DOWNLOAD_RESULT:')) {
          try {
            const jsonStr = message.split('DOWNLOAD_RESULT:')[1].split('MESSAGE_END')[0].trim()
            resolve(JSON.parse(jsonStr))
          } catch (err) {
            reject(err)
          } finally {
            botProcess.stdout?.off('data', handler)
          }
        }
      }
      botProcess.stdout?.on('data', handler)
    })
  }

  return {
    sendCommand,
    sendMessage: (jid, content) => 
      sendCommand(`SEND:${jid}|${formatContent(typeof content === 'string' ? content : '')}MESSAGE_END\n`, 'Write'),
    
    sendImage: (jid, image, caption = '', isUrl = false) => 
      sendCommand(createMediaCommand('IMAGE', jid, image, caption, isUrl), 'Image send'),
    
    sendVideo: (jid, video, caption = '', isUrl = false) => 
      sendCommand(createMediaCommand('VIDEO', jid, video, caption, isUrl), 'Video send'),
    
    sendAudio: (jid, audio, isUrl = false) => 
      sendCommand(createMediaCommand('AUDIO', jid, audio, '', isUrl), 'Audio send'),
    
    downloader: async (url, type, format) => {
      await sendCommand(`DOWNLOAD:${type}|${url}|${format || ''}MESSAGE_END\n`)
      return handleDownloadResponse()
    }
  }
}
