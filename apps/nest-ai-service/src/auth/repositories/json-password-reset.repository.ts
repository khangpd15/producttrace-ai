import { Injectable } from '@nestjs/common';
import * as fs from 'fs/promises';
import * as path from 'path';
import { PasswordResetRepository } from './password-reset.repository';

@Injectable()
export class JsonPasswordResetRepository implements PasswordResetRepository {
  private readonly filePath = path.join(process.cwd(), 'src', 'mock-data', 'password-reset-tokens.json');

  private async readData(): Promise<any[]> {
    try {
      const data = await fs.readFile(this.filePath, 'utf-8');
      return JSON.parse(data);
    } catch (error) {
      return [];
    }
  }

  private async writeData(data: any[]): Promise<void> {
    await fs.writeFile(this.filePath, JSON.stringify(data, null, 2), 'utf-8');
  }

  async saveToken(email: string, hashedToken: string, expiresAt: Date): Promise<void> {
    let tokens = await this.readData();
    // Remove existing token for the user if any
    tokens = tokens.filter((t) => t.email !== email);
    tokens.push({ email, hashedToken, expiresAt: expiresAt.toISOString() });
    await this.writeData(tokens);
  }

  async findByTokenAndEmail(hashedToken: string, email: string): Promise<any> {
    const tokens = await this.readData();
    return tokens.find((t) => t.hashedToken === hashedToken && t.email === email) || null;
  }

  async findByEmail(email: string): Promise<any> {
    const tokens = await this.readData();
    return tokens.find((t) => t.email === email) || null;
  }

  async deleteTokenByEmail(email: string): Promise<void> {
    let tokens = await this.readData();
    tokens = tokens.filter((t) => t.email !== email);
    await this.writeData(tokens);
  }

  async deleteExpiredTokens(): Promise<void> {
    let tokens = await this.readData();
    const now = new Date();
    tokens = tokens.filter((t) => new Date(t.expiresAt) > now);
    await this.writeData(tokens);
  }
}
