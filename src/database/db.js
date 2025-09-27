const { Pool } = require('pg');

class Database {
  constructor() {
    this.pool = new Pool({
      host: process.env.DB_HOST || 'localhost',
      port: process.env.DB_PORT || 5432,
      user: process.env.DB_USER || 'postgres',
      password: process.env.DB_PASSWORD || '',
      database: process.env.DB_NAME || 'loan_management',
      ssl: process.env.NODE_ENV === 'production' ? { rejectUnauthorized: false } : false,
    });

    this.pool.on('connect', () => {
      console.log('Connected to PostgreSQL database');
    });

    this.pool.on('error', (err) => {
      console.error('Database connection error:', err);
    });
  }

  async query(text, params) {
    const client = await this.pool.connect();
    try {
      const result = await client.query(text, params);
      return result;
    } catch (error) {
      console.error('Database query error:', error);
      throw error;
    } finally {
      client.release();
    }
  }

  async createTables() {
    try {
      // Users table
      await this.query(`
        CREATE TABLE IF NOT EXISTS users (
          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
          username VARCHAR(255) UNIQUE NOT NULL,
          password_hash VARCHAR(255) NOT NULL,
          full_name VARCHAR(255),
          created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
          updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
          deleted_at TIMESTAMP WITHOUT TIME ZONE
        )
      `);

      // Loans table
      await this.query(`
        CREATE TABLE IF NOT EXISTS loans (
          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
          user_id UUID REFERENCES users(id) NOT NULL,
          borrower_name VARCHAR(255) NOT NULL,
          amount NUMERIC NOT NULL,
          status VARCHAR(255) DEFAULT 'active',
          loan_date DATE NOT NULL,
          due_date DATE,
          created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
          updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
          deleted_at TIMESTAMP WITHOUT TIME ZONE
        )
      `);

      // Transactions table
      await this.query(`
        CREATE TABLE IF NOT EXISTS transactions (
          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
          loan_id UUID REFERENCES loans(id) NOT NULL,
          amount NUMERIC NOT NULL,
          remark TEXT,
          created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
          payment_date TIMESTAMP WITHOUT TIME ZONE,
          deleted_at TIMESTAMP WITHOUT TIME ZONE,
          updated_at TIMESTAMP WITHOUT TIME ZONE
        )
      `);

      console.log('Database tables created successfully');
    } catch (error) {
      console.error('Error creating tables:', error);
      throw error;
    }
  }

  async close() {
    await this.pool.end();
  }
}

module.exports = new Database();