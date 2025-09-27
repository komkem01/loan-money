const { validateJWT, extractTokenFromHeader } = require('../utils/jwt');
const { respondWithError } = require('../utils/response');
const db = require('../database/db');
const { User } = require('../models');

/**
 * Authentication middleware
 */
async function authMiddleware(req, res, next) {
  try {
    const authHeader = req.headers.authorization;
    
    if (!authHeader) {
      return respondWithError(res, 401, 'Authorization header required');
    }

    const token = extractTokenFromHeader(authHeader);
    const decoded = validateJWT(token);

    // Verify user still exists in database
    const result = await db.query(
      'SELECT * FROM users WHERE id = $1',
      [decoded.userId]
    );

    if (result.rows.length === 0) {
      return respondWithError(res, 401, 'User not found');
    }

    // Attach user to request
    const userData = result.rows[0];
    req.user = new User({
      id: userData.id,
      username: userData.username,
      email: userData.email,
      fullName: userData.full_name,
      phone: userData.phone,
      address: userData.address,
      createdAt: userData.created_at,
      updatedAt: userData.updated_at
    });

    next();
  } catch (error) {
    console.error('Auth middleware error:', error);
    return respondWithError(res, 401, 'Invalid or expired token');
  }
}

/**
 * Get user from request context
 */
function getUserFromContext(req) {
  return req.user;
}

module.exports = {
  authMiddleware,
  getUserFromContext
};