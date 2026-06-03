import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { getUser } from '../components/Navbar';

function Settings() {
  const navigate = useNavigate();

  useEffect(() => {
    const user = getUser();
    if (user) {
      navigate(`/user/${user.email}`, { replace: true });
    } else {
      navigate('/login', { replace: true });
    }
  }, [navigate]);

  return null;
}

export default Settings;
