import type { FunctionalComponent } from 'vue'
import { Braces, Fingerprint, QrCode } from 'lucide-vue-next'

export interface ToolNavItem {
  name: string
  path: string
  description: string
  icon: FunctionalComponent
}

export interface ToolGroup {
  label: string
  items: ToolNavItem[]
}

/**
 * Navigation catalogue, grouped the same way the backend categorises
 * tools (see /api/v1/tools).
 */
export const toolGroups: ToolGroup[] = [
  {
    label: 'Formatter',
    items: [
      {
        name: 'JSON Formatter',
        path: '/tools/json-formatter',
        description: 'Format, validate and minify JSON documents.',
        icon: Braces,
      },
    ],
  },
  {
    label: 'Generators',
    items: [
      {
        name: 'Hash Generator',
        path: '/tools/hash-generator',
        description: 'Generate SHA-256 and MD5 digests from text.',
        icon: Fingerprint,
      },
      {
        name: 'QR Generator',
        path: '/tools/qr-generator',
        description: 'Encode a URL or text as a QR code image.',
        icon: QrCode,
      },
    ],
  },
]
