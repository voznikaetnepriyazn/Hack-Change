import { describe, it, expect } from 'vitest';
import { isDatabaseStatus } from './databaseStatus';

describe('isDatabaseStatus', () => {
  it('should return true if the value is a valid database status', () => {
    expect(isDatabaseStatus('connected')).toBe(true);
  });
  it('should return true if the value is a valid database status', () => {
    expect(isDatabaseStatus('disconnected')).toBe(true);
  });
  it('should return true if the value is a valid database status', () => {
    expect(isDatabaseStatus('not_configured')).toBe(true);
  });
  it('should return false if the value is not a valid database status', () => {
    expect(isDatabaseStatus('invalid')).toBe(false);
  });
  it('should return false if the value is not a string', () => {
    expect(isDatabaseStatus(1)).toBe(false);
  });
  it('should return false if the value is not an object', () => {
    expect(isDatabaseStatus({})).toBe(false);
  });
});