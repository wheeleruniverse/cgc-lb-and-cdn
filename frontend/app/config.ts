/**
 * Application configuration based on deployment mode (full vs lite)
 *
 * This file centralizes feature flags and environment-based configuration,
 * demonstrating production-grade deployment practices with feature toggles.
 */

export type DeploymentMode = 'full' | 'lite'

export interface AppConfig {
  // Deployment mode
  deploymentMode: DeploymentMode
  isLiteMode: boolean
  isFullMode: boolean

  // Feature flags
  features: {
    enableAPI: boolean
    enableVoting: boolean
    enableCrossSessionTracking: boolean
    enableLiveStatistics: boolean
    enableImageGeneration: boolean
    useStaticData: boolean
  }

  // API configuration
  api: {
    baseUrl: string
  }

  // CDN configuration
  cdn: {
    spacesUrl: string
  }

  // Base path for GitHub Pages deployment
  basePath: string
}

/**
 * Get boolean value from environment variable
 */
function getEnvBoolean(key: string, defaultValue: boolean = false): boolean {
  const value = process.env[key]
  if (value === undefined) return defaultValue
  return value === 'true' || value === '1'
}

/**
 * Get string value from environment variable
 */
function getEnvString(key: string, defaultValue: string = ''): string {
  return process.env[key] || defaultValue
}

/**
 * Load and validate application configuration
 */
function loadConfig(): AppConfig {
  // Must access NEXT_PUBLIC_* env vars directly for browser build-time replacement
  // Using process.env[key] doesn't work in browser - Next.js only replaces direct access
  const deploymentMode = (process.env.NEXT_PUBLIC_DEPLOYMENT_MODE || 'lite') as DeploymentMode

  // Validate deployment mode - throw error instead of defaulting
  if (deploymentMode !== 'full' && deploymentMode !== 'lite') {
    throw new Error(
      `Invalid NEXT_PUBLIC_DEPLOYMENT_MODE: "${deploymentMode}". Must be "full" or "lite". ` +
      `Defaulting to "lite" to avoid costly full deployment errors.`
    )
  }

  const isLiteMode = deploymentMode === 'lite'
  const isFullMode = deploymentMode === 'full'

  // Core feature flag: enableAPI determines most other features
  const enableAPIEnv = process.env.NEXT_PUBLIC_ENABLE_API
  const enableAPI = enableAPIEnv === 'true' || enableAPIEnv === '1' || (!enableAPIEnv && isFullMode)

  return {
    deploymentMode,
    isLiteMode,
    isFullMode,

    features: {
      enableAPI,
      enableVoting: true, // Both modes support voting (backend vs localStorage)
      // These features are only available when API is enabled
      enableCrossSessionTracking: enableAPI,
      enableLiveStatistics: enableAPI,
      enableImageGeneration: enableAPI,
      useStaticData: !enableAPI, // Use static data when no API
    },

    api: {
      baseUrl: process.env.NEXT_PUBLIC_API_BASE_URL || '/api/v1',
    },

    cdn: {
      spacesUrl: process.env.NEXT_PUBLIC_SPACES_CDN_URL ||
        'https://cgc-lb-and-cdn-content.nyc3.cdn.digitaloceanspaces.com',
    },

    // Must access NEXT_PUBLIC_BASE_PATH directly for browser build-time replacement
    basePath: process.env.NEXT_PUBLIC_BASE_PATH || '',
  }
}

/**
 * Application configuration singleton
 */
export const config: AppConfig = loadConfig()

/**
 * Log configuration on app start (useful for debugging deployment issues)
 */
if (typeof window !== 'undefined') {
  console.log('[Config] Deployment mode:', config.deploymentMode)
  console.log('[Config] Feature flags:', config.features)
}

/**
 * Helper function to check if a feature is enabled
 */
export function isFeatureEnabled(feature: keyof AppConfig['features']): boolean {
  return config.features[feature]
}

/**
 * Get deployment mode display name
 */
export function getDeploymentDisplayName(): string {
  return config.isLiteMode
    ? 'Lite (GitHub Pages)'
    : 'Full (Digital Ocean)'
}
