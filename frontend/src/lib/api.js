const BASE = '/api'

async function req(method, path, body) {
  const res = await fetch(`${BASE}${path}`, {
    method,
    headers: body ? { 'Content-Type': 'application/json' } : {},
    body: body ? JSON.stringify(body) : undefined,
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(data.error || `Request failed (${res.status})`)
  return data
}

export const api = {
  getCycles:  ()           => req('GET',    '/cycles'),
  getCycle:   (id)         => req('GET',    `/cycles/${id}`),
  createCycle:(data)       => req('POST',   '/cycles', data),
  deleteCycle:(id)         => req('DELETE', `/cycles/${id}`),
  checkIn:    (id, data)   => req('POST',   `/periods/${id}/checkin`, data),
  undoCheckIn:(id)         => req('DELETE', `/periods/${id}/checkin`),
}
