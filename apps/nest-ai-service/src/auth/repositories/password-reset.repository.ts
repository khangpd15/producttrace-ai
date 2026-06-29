export interface PasswordResetRepository {
  saveToken(email: string, hashedToken: string, expiresAt: Date): Promise<void>;
  findByTokenAndEmail(hashedToken: string, email: string): Promise<any>;
  findByEmail(email: string): Promise<any>;
  deleteTokenByEmail(email: string): Promise<void>;
  deleteExpiredTokens(): Promise<void>;
}

export const PASSWORD_RESET_REPOSITORY = 'PASSWORD_RESET_REPOSITORY';
