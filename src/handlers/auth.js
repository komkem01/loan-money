const db = require('../database/db');
const { hashPassword, verifyPassword } = require('../utils/hash');
const { generateJWT } = require('../utils/jwt');
const { respondWithError, respondWithJSON, validateRequiredFields } = require('../utils/response');
const { User, AuthResponse } = require('../models');

class AuthHandler {
  /**
   * Register new user
   */
  async register(req, res) {
    try {
      const { username, password, fullName } = req.body;

      // Validate required fields
      validateRequiredFields(req.body, ['username', 'password']);

      // Validate password length
      if (password.length < 6) {
        return respondWithError(res, 400, 'Password must be at least 6 characters long');
      }

      // Check if user already exists
      const existingUser = await db.query(
        'SELECT id FROM users WHERE username = $1',
        [username]
      );

      if (existingUser.rows.length > 0) {
        return respondWithError(res, 400, 'Username already exists');
      }

      // Hash password
      const passwordHash = await hashPassword(password);

      // Create user
      const result = await db.query(
        `INSERT INTO users (username, password_hash, full_name)
         VALUES ($1, $2, $3)
         RETURNING *`,
        [username, passwordHash, fullName]
      );

      const userData = result.rows[0];
      const user = new User({
        id: userData.id,
        username: userData.username,
        fullName: userData.full_name,
        createdAt: userData.created_at,
        updatedAt: userData.updated_at
      });

      // Generate JWT token
      const token = generateJWT(user.id, user.username);

      return respondWithJSON(res, 201, new AuthResponse({ user, token }));

    } catch (error) {
      console.error('Register error:', error);
      return respondWithError(res, 500, 'Failed to register user');
    }
  }

  /**
   * Login user
   */
  async login(req, res) {
    try {
      const { username, password } = req.body;

      // Validate required fields
      validateRequiredFields(req.body, ['username', 'password']);

      // Find user by username
      const result = await db.query(
        'SELECT * FROM users WHERE username = $1',
        [username]
      );

      if (result.rows.length === 0) {
        return respondWithError(res, 401, 'Invalid credentials');
      }

      const userData = result.rows[0];

      // Verify password
      const isValidPassword = await verifyPassword(password, userData.password_hash);

      if (!isValidPassword) {
        return respondWithError(res, 401, 'Invalid credentials');
      }

      const user = new User({
        id: userData.id,
        username: userData.username,
        email: userData.email,
        fullName: userData.full_name,
        phone: userData.phone,
        address: userData.address,
        createdAt: userData.created_at,
        updatedAt: userData.updated_at
      });

      // Generate JWT token
      const token = generateJWT(user.id, user.username);

      return respondWithJSON(res, 200, new AuthResponse({ user, token }));

    } catch (error) {
      console.error('Login error:', error);
      return respondWithError(res, 500, 'Failed to login');
    }
  }

  /**
   * Get user from token (helper method)
   */
  async getUserFromToken(req) {
    try {
      const { validateJWT, extractTokenFromHeader } = require('../utils/jwt');
      
      const authHeader = req.headers.authorization;
      if (!authHeader) {
        throw new Error('No authorization header');
      }

      const token = extractTokenFromHeader(authHeader);
      const decoded = validateJWT(token);

      const result = await db.query(
        'SELECT * FROM users WHERE id = $1',
        [decoded.userId]
      );

      if (result.rows.length === 0) {
        throw new Error('User not found');
      }

      const userData = result.rows[0];
      return new User({
        id: userData.id,
        username: userData.username,
        email: userData.email,
        fullName: userData.full_name,
        phone: userData.phone,
        address: userData.address,
        createdAt: userData.created_at,
        updatedAt: userData.updated_at
      });

    } catch (error) {
      throw new Error('Failed to get user from token: ' + error.message);
    }
  }
}

module.exports = new AuthHandler();