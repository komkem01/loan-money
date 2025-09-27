const bcrypt = require('bcrypt');

const SALT_ROUNDS = 12;

/**
 * Hash password using bcrypt
 */
async function hashPassword(password) {
  try {
    return await bcrypt.hash(password, SALT_ROUNDS);
  } catch (error) {
    throw new Error('Failed to hash password');
  }
}

/**
 * Verify password against hash
 */
async function verifyPassword(password, hash) {
  try {
    return await bcrypt.compare(password, hash);
  } catch (error) {
    throw new Error('Failed to verify password');
  }
}

module.exports = {
  hashPassword,
  verifyPassword
};