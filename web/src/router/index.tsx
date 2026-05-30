import { createBrowserRouter } from 'react-router-dom';
import Home from '../pages/Home';
import FindPeople from '../pages/FindPeople';
import Posts from '../pages/Posts';

const router = createBrowserRouter([
  {
    path: '/',
    element: <Home />,
  },
  {
    path: '/find',
    element: <FindPeople />,
  },
  {
    path: '/posts',
    element: <Posts />,
  },
]);

export default router;
