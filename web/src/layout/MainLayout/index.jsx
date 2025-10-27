import { Outlet } from 'react-router-dom';
import AuthGuard from 'utils/route-guard/AuthGuard';

// material-ui
import { styled, useTheme } from '@mui/material/styles';
import { Box, CssBaseline } from '@mui/material';
import AdminContainer from 'ui-component/AdminContainer';

// project imports
import Breadcrumbs from 'ui-component/extended/Breadcrumbs';
import { AppSidebar } from '@/components/app-sidebar';
import { SidebarProvider, SidebarInset } from '@/components/ui/sidebar';
import navigation from 'menu-items';

// assets
import { Icon } from '@iconify/react';

// styles
export const Main = styled('main', { shouldForwardProp: (prop) => prop !== 'open' })(({ theme, open }) => ({
  ...theme.typography.mainContent,
  borderRadius: 0,
  backgroundColor: 'transparent',
  transition: theme.transitions.create(
    ['margin', 'width'],
    open
      ? {
          easing: theme.transitions.easing.easeOut,
          duration: theme.transitions.duration.enteringScreen
        }
      : {
          easing: theme.transitions.easing.sharp,
          duration: theme.transitions.duration.leavingScreen
        }
  ),
  overflowY: 'auto',
  overflowX: 'hidden',
  height: 'calc(100vh - 64px)',
  paddingBottom: '30px',
  position: 'relative',
  scrollbarWidth: 'thin',
  '&::-webkit-scrollbar': {
    width: '8px',
    height: '8px'
  },
  '&::-webkit-scrollbar-thumb': {
    background: theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.2)' : 'rgba(0, 0, 0, 0.15)',
    borderRadius: '4px'
  },
  '&::-webkit-scrollbar-track': {
    background: 'transparent'
  },
  [theme.breakpoints.up('md')]: {
    width: '100%',
    paddingLeft: theme.spacing(3),
    paddingRight: theme.spacing(3)
  },
  [theme.breakpoints.down('md')]: {
    marginLeft: '0',
    width: '100%',
    padding: '16px',
    height: 'calc(100vh - 64px)'
  },
  [theme.breakpoints.down('sm')]: {
    marginLeft: '0',
    width: '100%',
    padding: '16px',
    marginRight: '0',
    height: 'calc(100vh - 56px)'
  }
}));

// ==============================|| MAIN LAYOUT ||============================== //

const MainLayout = () => {
  return (
    <SidebarProvider>
      <Box
        sx={{
          display: 'flex',
          overflow: 'hidden',
          width: '100%',
          height: '100vh',
          position: 'relative'
        }}
      >
        <CssBaseline />

        {/* shadcn sidebar */}
        <AppSidebar />

        {/* main content area */}
        <SidebarInset>
          {/* main content */}
          <Main open={true}>
            {/* breadcrumb */}
            <Breadcrumbs separator={<Icon icon="solar:arrow-right-linear" width="16" />} navigation={navigation} icon title rightAlign />
            <AuthGuard>
              <AdminContainer>
                <Outlet />
              </AdminContainer>
            </AuthGuard>
          </Main>
        </SidebarInset>
      </Box>
    </SidebarProvider>
  );
};

export default MainLayout;
