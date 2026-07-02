import { Injectable } from '@nestjs/common';
import * as fs from 'fs/promises';
import * as path from 'path';
import { UserRepository } from './user.repository';

@Injectable()
export class JsonUserRepository implements UserRepository {
  private readonly filePath = path.join(process.cwd(), 'src', 'mock-data', 'users.json');

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

  async findByEmail(email: string): Promise<any> {
    const users = await this.readData();
    return users.find((user) => user.email === email) || null;
  }

  async updatePassword(userId: number, hashedPassword: string): Promise<void> {
    const users = await this.readData();
    const index = users.findIndex((u) => u.id === userId);
    if (index !== -1) {
      users[index].password = hashedPassword;
      await this.writeData(users);
    }
  }
}
