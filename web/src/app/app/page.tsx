"use client"

import { useState } from 'react'
import { requestOTP, verifyOTP } from '@/src/lib/api'
import { getOrCreateDeviceId } from '@/src/lib/device'

export default function AppPage() {
  const [step, setStep] = useState<'phone' | 'otp' | 'done'>('phone')
  const [phone, setPhone] = useState('')
  const [code, setCode] = useState('')
  const [token, setToken] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  async function onRequestOTP() {
    setError('')
    setLoading(true)
    try {
      const device_id = getOrCreateDeviceId()
      await requestOTP(phone, device_id)
      setStep('otp')
    } catch (e: any) {
      setError(e.message || 'Failed to request OTP')
    } finally {
      setLoading(false)
    }
  }

  async function onVerify() {
    setError('')
    setLoading(true)
    try {
      const device_id = getOrCreateDeviceId()
      const res = await verifyOTP(phone, code, device_id)
      setToken(res.token)
      setStep('done')
    } catch (e: any) {
      setError(e.message || 'Failed to verify OTP')
    } finally {
      setLoading(false)
    }
  }

  return (
    <main className="container-prose py-24">
      <div className="max-w-md mx-auto card">
        <h1 className="text-2xl font-semibold">Start your order</h1>
        <p className="mt-2 text-white/70">Sign in with your phone number.</p>

        {step === 'phone' && (
          <div className="mt-6 space-y-4">
            <input
              type="tel"
              placeholder="Phone number"
              className="w-full rounded-md bg-white/5 border border-white/10 px-4 py-3 outline-none focus:ring-2 focus:ring-brand"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
            />
            <button className="btn-primary w-full" onClick={onRequestOTP} disabled={loading || phone.length < 6}>
              {loading ? 'Sending…' : 'Send OTP'}
            </button>
          </div>
        )}

        {step === 'otp' && (
          <div className="mt-6 space-y-4">
            <input
              type="text"
              placeholder="Enter OTP"
              className="w-full rounded-md bg-white/5 border border-white/10 px-4 py-3 outline-none focus:ring-2 focus:ring-brand tracking-widest"
              value={code}
              onChange={(e) => setCode(e.target.value)}
            />
            <button className="btn-primary w-full" onClick={onVerify} disabled={loading || code.length < 4}>
              {loading ? 'Verifying…' : 'Verify'}
            </button>
            <button className="w-full text-sm text-white/70 hover:text-white" onClick={() => setStep('phone')}>Use a different number</button>
          </div>
        )}

        {step === 'done' && (
          <div className="mt-6 space-y-3">
            <p className="text-white/80">You are signed in.</p>
            <div className="text-xs break-all text-white/60">Token: {token}</div>
            <div className="mt-6 grid gap-3 md:grid-cols-2">
              <div className="card">Your orders will appear here.</div>
              <div className="card">Explore the menu and add items.</div>
            </div>
          </div>
        )}

        {error && <p className="mt-4 text-red-400 text-sm">{error}</p>}
      </div>
    </main>
  )
}



