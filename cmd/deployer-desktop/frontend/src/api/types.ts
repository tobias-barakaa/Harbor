export interface ApplicationEntry {
  id: string
  path: string
  name: string
  runtime: string
  framework: string
  strategy: string
  port: number
  addedAt: string
}

export interface ServerInfo {
  name: string
  host: string
  port: number
  username: string
  authMethod: string
}

export interface AddServerInput {
  name: string
  host: string
  port: number
  username: string
  authMethod: 'private_key' | 'password'
  privateKeyPath: string
  password: string
}

export interface PortMappingInfo {
  host: number
  container: number
}

export interface DeploymentInfo {
  name: string
  serverName: string
  containerName: string
  image: string
  ports: PortMappingInfo[]
  deployedAt: string
  status: string
}

export interface DeployInput {
  applicationId: string
  serverName: string
}

export type ErrorCategory = 'validation' | 'connection' | 'inspection' | 'build' | 'deployment' | 'unexpected'

export interface CategorizedError {
  category: ErrorCategory
  message: string
}

// Go errors are thrown as plain strings by Wails. Every categorized
// error on the Go side is formatted "<category>: <message>" (see
// cmd/deployer-desktop/errors.go) — this splits that back apart.
// Anything that doesn't match (a raw JS error, say) falls back to
// "unexpected" rather than losing the message.
export function parseError(err: unknown): CategorizedError {
  const raw = err instanceof Error ? err.message : String(err)
  const known: ErrorCategory[] = ['validation', 'connection', 'inspection', 'build', 'deployment', 'unexpected']
  const sep = raw.indexOf(': ')
  if (sep > 0) {
    const cat = raw.slice(0, sep)
    if ((known as string[]).includes(cat)) {
      return { category: cat as ErrorCategory, message: raw.slice(sep + 2) }
    }
  }
  return { category: 'unexpected', message: raw }
}