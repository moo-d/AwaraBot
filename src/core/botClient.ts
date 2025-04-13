import { ChildProcess } from 'child_process'
import { AIResponse, Bot } from '../types'

export function createBotClient(botProcess: ChildProcess): Bot {
  const formatContent = (content: string) => content.replace(/\n/g, '{{NL}}')

  const sendCommand = (command: string, errorPrefix = 'Command') => {
    return new Promise<void>((resolve, reject) => {
      botProcess.stdin?.write(command, err => {
        err ? reject(`${errorPrefix} error: ${err}`) : resolve()
      })
    })
  }

  const createMediaCommand = (
    type: 'IMAGE' | 'VIDEO' | 'AUDIO',
    jid: string,
    media: string | Buffer,
    caption = '',
    isUrl = false
  ) => {
    const baseCmd = isUrl || typeof media === 'string' ? `SEND_URL_${type}` : `SEND_${type}`
    const mediaData = typeof media === 'string' ? media : media.toString('base64')
    return `${baseCmd}:${jid}|${mediaData}${type !== 'AUDIO' ? `|${formatContent(caption)}` : ''}MESSAGE_END\n`
  }

  const handleResponse = (prefix: string): Promise<any> => {
    return new Promise((resolve, reject) => {
      const handler = (data: Buffer) => {
        const message = data.toString()
        if (message.includes(`${prefix}:`)) {
          try {
            const jsonStr = message.split(`${prefix}:`)[1].split('MESSAGE_END')[0].trim()
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

  const ai = async (
    jid: string,
    prompt: string,
    messages: { role: string, content: string }[] = [],
    model = "GPT-4"
  ): Promise<AIResponse> => {
    const messagesStr = JSON.stringify(messages)
    const command = `CHATBOT:${jid}|${prompt}|${model}|${messagesStr}MESSAGE_END\n`
    
    await sendCommand(command, 'Chatbot')
    
    return new Promise((resolve) => {
      const handler = (data: Buffer) => {
        const message = data.toString()
        if (message.includes('CHATBOT_RESULT:')) {
          try {
            const jsonStr = message.split('CHATBOT_RESULT:')[1].split('MESSAGE_END')[0].trim()
            const result = JSON.parse(jsonStr)
            
            resolve({
              chat: result.chat || jid,
              message: result.message || result.caption || 'No response',
              command: result.command,
              query: result.query,
              caption: result.caption
            })
            
          } catch (err) {
            console.error('Failed to parse chatbot response:', err)
            resolve({
              chat: jid,
              message: 'Error processing response'
            })
          } finally {
            botProcess.stdout?.off('data', handler)
          }
        }
      }
      botProcess.stdout?.on('data', handler)
    })
  }

  return {
    ai,
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
      return handleResponse('DOWNLOAD_RESULT')
    },
    
    sendReaction: (jid, sender, messageId, emoji) => {
      const command = `REACT:${jid}|${messageId}|${formatContent(emoji)}|${sender}MESSAGE_END\n`
      return sendCommand(command, 'Reaction')
    }
  }
}