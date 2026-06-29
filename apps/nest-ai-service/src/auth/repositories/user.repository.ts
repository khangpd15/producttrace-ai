export interface UserRepository {
  findByEmail(email: string): Promise<any>;
  updatePassword(userId: number, hashedPassword: string): Promise<void>;
}

export const USER_REPOSITORY = 'USER_REPOSITORY';
