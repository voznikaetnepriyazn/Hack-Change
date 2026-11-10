export const validateRegisterData = (email: string, password: string) => {
  email = email.trim();

  if (!email || !password) {
    throw new Error("Password and email are required");
  }

  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
    throw new Error('Email must be a valid email address');
  }

  if (password.length > 100 || password.length < 8) {
    throw new Error('Password must be between 8 and 100 characters');
  }

  return { email, password };
}
export const validateLoginData = (email: string, password: string) => {
  email = email.trim();

  if (!email || !password) {
    throw new Error("Password and email are required");
  }

  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
    throw new Error('Email must be a valid email address');
  }

  if (password.length > 100 || password.length < 8) {
    throw new Error('Password must be between 8 and 100 characters');
  }

  return { email, password };
}