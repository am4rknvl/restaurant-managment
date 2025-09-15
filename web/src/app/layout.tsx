import './globals.css'
import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Restaurant — Seamless Table Ordering',
  description: 'Order from your table. Fast, simple, secure.',
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  )
}



