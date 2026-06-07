import { createBrowserRouter } from 'react-router-dom';
import App from '../App';
import Home from '../pages/Home';
import FindPeople from '../pages/FindPeople';
import UserDetail from '../pages/UserDetail';
import Posts from '../pages/Posts';
import Login from '../pages/Login';
import Register from '../pages/Register';
import ForgotPassword from '../pages/ForgotPassword';
import Appointments from '../pages/Appointments';
import Settings from '../pages/Settings';
import CreateCard from '../pages/CreateCard';
import AdminLayout from '../pages/admin/AdminLayout';
import AdminDashboard from '../pages/admin/AdminDashboard';
import AdminUsers from '../pages/admin/AdminUsers';
import AdminCards from '../pages/admin/AdminCards';
import AdminAppointments from '../pages/admin/AdminAppointments';

const router = createBrowserRouter([
  {
    element: <App />,
    children: [
      { path: '/', element: <Home /> },
      { path: '/find', element: <FindPeople /> },
      { path: '/user/:id', element: <UserDetail /> },
      { path: '/posts', element: <Posts /> },
      { path: '/login', element: <Login /> },
      { path: '/register', element: <Register /> },
      { path: '/forgot-password', element: <ForgotPassword /> },
      { path: '/appointments', element: <Appointments /> },
      { path: '/settings', element: <Settings /> },
      { path: '/my-card', element: <CreateCard /> },
    ],
  },
  {
    element: <AdminLayout />,
    children: [
      { path: '/admin', element: <AdminDashboard /> },
      { path: '/admin/users', element: <AdminUsers /> },
      { path: '/admin/cards', element: <AdminCards /> },
      { path: '/admin/appointments', element: <AdminAppointments /> },
    ],
  },
]);

export default router;
