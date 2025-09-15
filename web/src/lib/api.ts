const API_BASE = process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:8080/api/v1'

export async function requestOTP(phone_number: string, device_id: string) {
  const res = await fetch(`${API_BASE}/auth/request-otp`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ phone_number, device_id }),
    cache: 'no-store',
  })
  if (!res.ok) throw new Error(await res.text())
  return res.json()
}

export async function verifyOTP(phone_number: string, code: string, device_id: string) {
  const res = await fetch(`${API_BASE}/auth/verify-otp`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ phone_number, code, device_id }),
    cache: 'no-store',
  })
  if (!res.ok) throw new Error(await res.text())
  return res.json() as Promise<{ token: string }>
}



