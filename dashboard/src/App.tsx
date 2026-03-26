import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import { Layout } from '@/components/layout/Layout'
import { Overview } from '@/pages/Overview'
import { Keys } from '@/pages/Keys'
import { Analytics } from '@/pages/Analytics'
import { Models } from '@/pages/Models'
import { Routing } from '@/pages/Routing'
import { Cache } from '@/pages/Cache'
import { Alerts } from '@/pages/Alerts'
import { Traces } from '@/pages/Traces'
import { Settings } from '@/pages/Settings'

const router = createBrowserRouter([
  {
    path: '/',
    element: <Layout />,
    children: [
      { index: true, element: <Overview /> },
      { path: 'keys', element: <Keys /> },
      { path: 'analytics', element: <Analytics /> },
      { path: 'models', element: <Models /> },
      { path: 'routing', element: <Routing /> },
      { path: 'cache', element: <Cache /> },
      { path: 'alerts', element: <Alerts /> },
      { path: 'traces', element: <Traces /> },
      { path: 'settings', element: <Settings /> },
    ],
  },
])

export function App() {
  return <RouterProvider router={router} />
}
