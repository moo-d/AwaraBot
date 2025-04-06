import { Command } from '../types'

export default {
  name: 'tiktok',
  alias: ['tt', 'tiktokdl'],
  category: 'downloader',
  description: 'Download TikTok video without watermark',
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
      await bot.sendMessage(context.chat, '⏳ Downloading TikTok video...')
      
      const response = await bot.downloader(url, 'tiktok')
      
      if (!response?.result?.video) {
        throw new Error(response?.error || 'No video found in response')
      }
  
      await bot.sendVideo(context.chat, response.result.video, 'TikTok Video', true)
      
      if (response.result.music) {
        await bot.sendAudio(context.chat, response.result.music, true)
      }
    } catch (error: any) {
      await bot.sendMessage(
        context.chat, 
        `❌ Failed to download TikTok video: ${error.message}`
      )
    }
  }
} as Command