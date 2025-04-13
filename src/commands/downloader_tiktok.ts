import { Command } from '../types'

export default {
  name: 'tiktok',
  alias: ['tt', 'tiktokdl'],
  category: 'downloader',
  wait: true,
  description: 'Download TikTok video or images without watermark',
  async handler(bot, args, context) {
    if (!args.length) {
      return bot.sendMessage(context.chat, 
        '⚠️ Please provide a TikTok URL\nExample: /tiktok https://vm.tiktok.com/xyz'
      )
    }

    const url = args[0]
    if (!url.match(/tiktok\.com|vm\.tiktok\.com|vt\.tiktok\.com/)) {
      return bot.sendMessage(context.chat, 
        '❌ Invalid TikTok URL. Please provide a valid TikTok link.'
      )
    }

    try {
      const { result, error } = await bot.downloader(url, 'tiktok')
      
      if (!result) {
        throw new Error(error || 'No content found in response')
      }

      const sender = context.from.split(':')[0] + '@s.whatsapp.net'
      const isGroup = context.isGroup
      const sendOperations: Promise<unknown>[] = []

      if (result.images?.length) {
        const sendPrivate = isGroup && result.images.length > 1
        
        if (sendPrivate) {
          await bot.sendMessage(context.chat, 
            `📸 Found ${result.images.length} images. Sending to your private chat.`
          )
        }

        const targetChat = sendPrivate ? sender : context.chat
        const musicTarget = sendPrivate ? sender : context.chat

        for (const image of result.images) {
          sendOperations.push(bot.sendImage(targetChat, image))
        }

        if (result.music) {
          sendOperations.push(bot.sendAudio(musicTarget, result.music, !sendPrivate))
        }
      } else if (result.video) {
        sendOperations.push(
          bot.sendVideo(context.chat, result.video, 'TikTok Video', true)
        )
        
        if (result.music) {
          sendOperations.push(
            bot.sendAudio(context.chat, result.music, true)
          )
        }
      } else {
        throw new Error('No video or images found in response')
      }

      await Promise.all(sendOperations)

    } catch (error) {
      const errorMessage = error instanceof Error 
        ? error.message 
        : 'An unknown error occurred'
      await bot.sendMessage(
        context.chat, 
        `❌ Failed to download TikTok content: ${errorMessage}`
      )
    }
  }
} as Command