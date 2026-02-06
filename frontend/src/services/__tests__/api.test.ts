import { describe, it, expect, vi, beforeEach } from 'vitest'
import { api } from '../api'

describe('ApiClient', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    window.localStorage.getItem = vi.fn().mockReturnValue(null)
  })

  describe('get', () => {
    it('makes GET request to correct URL', async () => {
      const mockResponse = { data: 'test' }
      vi.mocked(global.fetch).mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      } as Response)

      const result = await api.get('/test')

      expect(global.fetch).toHaveBeenCalledWith(
        '/api/v1/test',
        expect.objectContaining({
          method: 'GET',
          headers: expect.objectContaining({
            'Content-Type': 'application/json',
          }),
        })
      )
      expect(result).toEqual(mockResponse)
    })

    it('includes auth token when present', async () => {
      window.localStorage.getItem = vi.fn().mockReturnValue('test-token')
      vi.mocked(global.fetch).mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({}),
      } as Response)

      await api.get('/test')

      expect(global.fetch).toHaveBeenCalledWith(
        '/api/v1/test',
        expect.objectContaining({
          headers: expect.objectContaining({
            'Authorization': 'Bearer test-token',
          }),
        })
      )
    })
  })

  describe('post', () => {
    it('makes POST request with JSON body', async () => {
      const mockData = { username: 'test' }
      vi.mocked(global.fetch).mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ success: true }),
      } as Response)

      await api.post('/test', mockData)

      expect(global.fetch).toHaveBeenCalledWith(
        '/api/v1/test',
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify(mockData),
        })
      )
    })
  })

  describe('error handling', () => {
    it('throws error with message from response', async () => {
      vi.mocked(global.fetch).mockResolvedValue({
        ok: false,
        json: () => Promise.resolve({ error: 'Custom error message' }),
      } as Response)

      await expect(api.get('/test')).rejects.toThrow('Custom error message')
    })

    it('throws generic error when response has no error field', async () => {
      vi.mocked(global.fetch).mockResolvedValue({
        ok: false,
        json: () => Promise.resolve({}),
      } as Response)

      await expect(api.get('/test')).rejects.toThrow('Request failed')
    })

    it('throws generic error when response is not JSON', async () => {
      vi.mocked(global.fetch).mockResolvedValue({
        ok: false,
        json: () => Promise.reject(new Error('Invalid JSON')),
      } as Response)

      await expect(api.get('/test')).rejects.toThrow('Request failed')
    })
  })

  describe('put', () => {
    it('makes PUT request with JSON body', async () => {
      const mockData = { name: 'updated' }
      vi.mocked(global.fetch).mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ success: true }),
      } as Response)

      await api.put('/test/1', mockData)

      expect(global.fetch).toHaveBeenCalledWith(
        '/api/v1/test/1',
        expect.objectContaining({
          method: 'PUT',
          body: JSON.stringify(mockData),
        })
      )
    })
  })

  describe('delete', () => {
    it('makes DELETE request', async () => {
      vi.mocked(global.fetch).mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ success: true }),
      } as Response)

      await api.delete('/test/1')

      expect(global.fetch).toHaveBeenCalledWith(
        '/api/v1/test/1',
        expect.objectContaining({
          method: 'DELETE',
        })
      )
    })
  })
})
