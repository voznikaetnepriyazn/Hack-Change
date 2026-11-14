export const validateLoginData = (login: string, password: string) => {
  login = login.trim();

  if (!login || !password) {
    throw new Error("Password and login are required");
  }

  if (password.length > 100 || password.length < 8) {
    throw new Error('Password must be between 8 and 100 characters');
  }

  return { login, password };
}

