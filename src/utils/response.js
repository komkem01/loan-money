/**
 * Send error response
 */
function respondWithError(res, code, message) {
  return res.status(code).json({
    error: {
      message: message || 'An error occurred',
      status: code
    }
  });
}

/**
 * Send success response with data
 */
function respondWithJSON(res, code, data) {
  return res.status(code).json({
    data: data,
    success: true,
    status: code
  });
}

/**
 * Send success response with message
 */
function respondWithMessage(res, code, message) {
  return res.status(code).json({
    message: message,
    success: true,
    status: code
  });
}

/**
 * Log database error
 */
function logDatabaseError(operation, error) {
  console.error(`Database ${operation} error:`, {
    message: error.message,
    stack: error.stack,
    timestamp: new Date().toISOString()
  });
}

/**
 * Log API call
 */
function logAPICall(method, path, userId, statusCode) {
  console.log(`[${new Date().toISOString()}] ${method} ${path} - User: ${userId || 'anonymous'} - Status: ${statusCode}`);
}

/**
 * Validate required fields
 */
function validateRequiredFields(data, requiredFields) {
  const missingFields = [];
  
  requiredFields.forEach(field => {
    if (!data[field] || (typeof data[field] === 'string' && data[field].trim() === '')) {
      missingFields.push(field);
    }
  });
  
  if (missingFields.length > 0) {
    throw new Error(`Missing required fields: ${missingFields.join(', ')}`);
  }
}

/**
 * Parse pagination parameters
 */
function parsePagination(query) {
  const page = parseInt(query.page) || 1;
  const limit = parseInt(query.limit) || 10;
  const offset = (page - 1) * limit;
  
  return { page, limit, offset };
}

/**
 * Format currency
 */
function formatCurrency(amount) {
  return new Intl.NumberFormat('th-TH', {
    style: 'currency',
    currency: 'THB'
  }).format(amount);
}

module.exports = {
  respondWithError,
  respondWithJSON,
  respondWithMessage,
  logDatabaseError,
  logAPICall,
  validateRequiredFields,
  parsePagination,
  formatCurrency
};