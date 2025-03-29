const PREFIX_REGEX = /^[°•π÷×¶∆£¢€¥®™✓=|~zZ+×_*!#%^&./\\©^]/

export function extractCommand(input: string): {
  prefix: string | null
  command: string
  args: string[]
} {
  const trimmedInput = input.trim()
  if (!trimmedInput) return { prefix: null, command: '', args: [] }

  const [fullCommand, ...rawArgs] = trimmedInput.split(/\s+/)
  const args = rawArgs.filter(Boolean)
  
  const prefix = PREFIX_REGEX.test(fullCommand) ? fullCommand[0] : null
  const command = prefix ? fullCommand.slice(1) : fullCommand

  return { prefix, command, args }
}
