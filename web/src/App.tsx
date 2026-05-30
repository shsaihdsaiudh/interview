import { Outlet } from 'react-router-dom';
import Navbar from './components/Navbar';

// App 作为全局布局组件，包含导航栏和页面内容
export default function App() {
  return (
    <>
      <Navbar />
      <Outlet />
    </>
  );
}
