import { describe, expect, it } from 'vitest'
import { ApiError, errorMessage, getApiErrorInfo, redactSensitive } from './api'

describe('api error mapping', () => {
  it('maps forbidden errors to permission copy with metadata', () => {
    const err = new ApiError('connection is read-only', {
      status: 403,
      code: 'FORBIDDEN',
      category: 'auth_error',
      type: 'forbidden_request',
      requestId: 'req-123',
    })

    expect(getApiErrorInfo(err)).toEqual({
      title: 'Permission denied',
      message: 'connection is read-only',
      detail: 'code: FORBIDDEN · request: req-123',
      status: 403,
      code: 'FORBIDDEN',
      requestId: 'req-123',
      retryable: false,
    })
  })

  it('marks timeout and server errors as retryable', () => {
    expect(getApiErrorInfo(new ApiError('slow', { status: 408 })).retryable).toBe(true)
    expect(getApiErrorInfo(new ApiError('boom', { status: 503 })).retryable).toBe(true)
  })

  it('redacts secrets from messages and details', () => {
    const err = new ApiError(
      'connect mongodb://admin:secret@localhost/test?api_key=abc',
      {
        details: {
          password: 'secret',
          uri: 'mysql://root:pw@localhost/db',
        },
      },
    )

    expect(err.message).not.toContain('secret')
    expect(err.message).not.toContain('abc')
    expect(err.details?.password).toBe('***')
    expect(err.details?.uri).toBe('mysql://***:***@localhost/db')
  })

  it('formats display messages with machine details', () => {
    const err = new ApiError('blocked', { code: 'READONLY_VIOLATION' })

    expect(errorMessage(err)).toBe('blocked (code: READONLY_VIOLATION)')
  })

  it('redacts authorization headers', () => {
    expect(redactSensitive('Authorization: Bearer abc.def.ghi')).toBe('Authorization: Bearer ***')
  })
})
