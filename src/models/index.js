// User model
class User {
  constructor({
    id = null,
    username,
    passwordHash,
    fullName = null,
    createdAt = new Date(),
    updatedAt = new Date(),
    deletedAt = null
  }) {
    this.id = id;
    this.username = username;
    this.passwordHash = passwordHash;
    this.fullName = fullName;
    this.createdAt = createdAt;
    this.updatedAt = updatedAt;
    this.deletedAt = deletedAt;
  }

  toJSON() {
    const { passwordHash, ...userWithoutPassword } = this;
    return userWithoutPassword;
  }
}

// Loan model
class Loan {
  constructor({
    id = null,
    userId,
    borrowerName,
    borrowerPhone = null,
    borrowerAddress = null,
    amount,
    interestRate,
    loanDate,
    dueDate,
    status = 'active',
    notes = null,
    createdAt = new Date(),
    updatedAt = new Date()
  }) {
    this.id = id;
    this.userId = userId;
    this.borrowerName = borrowerName;
    this.borrowerPhone = borrowerPhone;
    this.borrowerAddress = borrowerAddress;
    this.amount = parseFloat(amount);
    this.interestRate = parseFloat(interestRate);
    this.loanDate = loanDate;
    this.dueDate = dueDate;
    this.status = status;
    this.notes = notes;
    this.createdAt = createdAt;
    this.updatedAt = updatedAt;
  }
}

// Transaction model
class Transaction {
  constructor({
    id = null,
    loanId,
    userId,
    amount,
    transactionType,
    transactionDate,
    description = null,
    createdAt = new Date(),
    updatedAt = new Date()
  }) {
    this.id = id;
    this.loanId = loanId;
    this.userId = userId;
    this.amount = parseFloat(amount);
    this.transactionType = transactionType;
    this.transactionDate = transactionDate;
    this.description = description;
    this.createdAt = createdAt;
    this.updatedAt = updatedAt;
  }
}

// Request/Response DTOs
class AuthRequest {
  constructor({ username, password, email, fullName = null }) {
    this.username = username;
    this.password = password;
    this.email = email;
    this.fullName = fullName;
  }
}

class LoginRequest {
  constructor({ username, password }) {
    this.username = username;
    this.password = password;
  }
}

class LoanCreateRequest {
  constructor({
    borrowerName,
    borrowerPhone,
    borrowerAddress,
    amount,
    interestRate,
    loanDate,
    dueDate,
    notes
  }) {
    this.borrowerName = borrowerName;
    this.borrowerPhone = borrowerPhone;
    this.borrowerAddress = borrowerAddress;
    this.amount = amount;
    this.interestRate = interestRate;
    this.loanDate = loanDate;
    this.dueDate = dueDate;
    this.notes = notes;
  }
}

class TransactionCreateRequest {
  constructor({
    loanId,
    amount,
    transactionType,
    transactionDate,
    description
  }) {
    this.loanId = loanId;
    this.amount = amount;
    this.transactionType = transactionType;
    this.transactionDate = transactionDate;
    this.description = description;
  }
}

// Response DTOs
class AuthResponse {
  constructor({ user, token }) {
    this.user = user.toJSON ? user.toJSON() : user;
    this.token = token;
  }
}

class DashboardStats {
  constructor({
    totalLoans = 0,
    activeLoans = 0,
    totalAmount = 0,
    totalInterest = 0,
    overdueLoans = 0,
    recentTransactions = []
  }) {
    this.totalLoans = totalLoans;
    this.activeLoans = activeLoans;
    this.totalAmount = totalAmount;
    this.totalInterest = totalInterest;
    this.overdueLoans = overdueLoans;
    this.recentTransactions = recentTransactions;
  }
}

module.exports = {
  User,
  Loan,
  Transaction,
  AuthRequest,
  LoginRequest,
  LoanCreateRequest,
  TransactionCreateRequest,
  AuthResponse,
  DashboardStats
};