const db = require('../database/db');
const { respondWithError, respondWithJSON, validateRequiredFields } = require('../utils/response');
const { getUserFromContext } = require('../middleware/auth');
const { hashPassword, verifyPassword } = require('../utils/hash');

class ProfileHandler {
  /**
   * Get user profile
   */
  async getProfile(req, res) {
    try {
      const user = getUserFromContext(req);
      return respondWithJSON(res, 200, user.toJSON());
    } catch (error) {
      console.error('Get profile error:', error);
      return respondWithError(res, 500, 'Failed to get profile');
    }
  }

  /**
   * Update user profile
   */
  async updateProfile(req, res) {
    try {
      const user = getUserFromContext(req);
      const { fullName, phone, address, email } = req.body;

      const result = await db.query(
        `UPDATE users 
         SET full_name = $1, phone = $2, address = $3, email = $4, updated_at = CURRENT_TIMESTAMP
         WHERE id = $5
         RETURNING *`,
        [fullName, phone, address, email, user.id]
      );

      if (result.rows.length === 0) {
        return respondWithError(res, 404, 'User not found');
      }

      const updatedUserData = result.rows[0];
      const { User } = require('../models');
      const updatedUser = new User({
        id: updatedUserData.id,
        username: updatedUserData.username,
        email: updatedUserData.email,
        fullName: updatedUserData.full_name,
        phone: updatedUserData.phone,
        address: updatedUserData.address,
        createdAt: updatedUserData.created_at,
        updatedAt: updatedUserData.updated_at
      });

      return respondWithJSON(res, 200, updatedUser.toJSON());

    } catch (error) {
      console.error('Update profile error:', error);
      return respondWithError(res, 500, 'Failed to update profile');
    }
  }

  /**
   * Change password
   */
  async changePassword(req, res) {
    try {
      const user = getUserFromContext(req);
      const { currentPassword, newPassword } = req.body;

      validateRequiredFields(req.body, ['currentPassword', 'newPassword']);

      if (newPassword.length < 6) {
        return respondWithError(res, 400, 'New password must be at least 6 characters long');
      }

      // Get current password hash
      const result = await db.query(
        'SELECT password_hash FROM users WHERE id = $1',
        [user.id]
      );

      if (result.rows.length === 0) {
        return respondWithError(res, 404, 'User not found');
      }

      // Verify current password
      const isValidPassword = await verifyPassword(currentPassword, result.rows[0].password_hash);
      if (!isValidPassword) {
        return respondWithError(res, 400, 'Current password is incorrect');
      }

      // Hash new password
      const newPasswordHash = await hashPassword(newPassword);

      // Update password
      await db.query(
        'UPDATE users SET password_hash = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2',
        [newPasswordHash, user.id]
      );

      return respondWithJSON(res, 200, { message: 'Password changed successfully' });

    } catch (error) {
      console.error('Change password error:', error);
      return respondWithError(res, 500, 'Failed to change password');
    }
  }
}

module.exports = new ProfileHandler();