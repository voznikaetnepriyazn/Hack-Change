import { useState, type ComponentProps } from "react";
import { Link, useNavigate } from "react-router-dom";
import ApiError from "../../shared/errors/ApiError";
import { useAuth } from "../context/AuthContext";

export const LoginPage = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  
  const { login } = useAuth();
  const navigate = useNavigate();
  
  const handleSubmit:NonNullable<ComponentProps<'form'>['onSubmit']> = async (e) =>{
    e.preventDefault();
    setErrorMessage(null);
    try {
      await login(email, password);
      navigate('/account');
    } catch (error) {
      if (error instanceof ApiError) {
        setErrorMessage(error.message);
      } else {
        console.error(error);
        setErrorMessage('An error occurred while logging in. Please try again.');
      }
    }
  }

  return (
    <main className="page">
      <h1>Login</h1>
      <form onSubmit={handleSubmit}>
        <label>
          Email
          <input type="email" id="email" name="email" required value={email} onChange={(e) => setEmail(e.target.value)} />
        </label>
        <label>
          Password
          <input type="password" id="password" name="password" required value={password} onChange={(e) => setPassword(e.target.value)} />
        </label>
        <button type="submit">Login</button>
        {errorMessage && <p className="error">{errorMessage}</p>}
      </form>
      <p>Don't have an account? <Link to="/auth/register">Register</Link></p>
    </main>
  )
}