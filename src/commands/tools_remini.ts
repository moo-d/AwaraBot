// import { Command } from '../types'

// export default {
//   name: 'remini',
//   alias: ['hd', 'hdr', 'enhance'],
//   category: 'tools',
//   description: 'Enhance image quality using AI',
//   wait: true,
//   async handler(bot, args, context) {
//     try {
//       // Validate we have an image to process
//       if (!context.isImage && !context.isQuotedImage) {
//         throw new Error('Please send or reply to an image\nExample: /remini [reply to image]')
//       }

//       await bot.sendReaction(context.chat, context.sender, context.messageId, '⏳')

//       // Get the correct message details
//       const isQuoted = context.isQuotedImage
//       const targetMessageId = isQuoted 
//         ? context.quotedMessage?.messageId 
//         : context.messageId

//       if (!targetMessageId) {
//         throw new Error('Could not find image message ID')
//       }

//       // Download the media
//       const imageBuffer = await bot.downloadMedia(
//         targetMessageId,
//         context.chat,
//         isQuoted ? 'quoted' : 'direct'
//       )

//       if (!imageBuffer) {
//         throw new Error('Failed to download the image')
//       }

//       // Enhance the image
//       const { result, error } = await bot.enhanceImage(imageBuffer, 'enhance')
//       if (!result) {
//         throw new Error(error || 'Failed to enhance image')
//       }

//       await bot.sendImage(context.chat, result.url, 'Here\'s your enhanced image', true)
//       await bot.sendReaction(context.chat, context.sender, context.messageId, '✅')

//     } catch (error) {
//       console.error('[REMINI] Error:', error)
//       await bot.sendMessage(
//         context.chat,
//         `❌ ${error instanceof Error ? error.message : 'Failed to enhance image'}`
//       )
//       await bot.sendReaction(context.chat, context.sender, context.messageId, '❌')
//     }
//   }
// } as Command