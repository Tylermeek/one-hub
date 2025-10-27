import PropTypes from 'prop-types';
import { useLocation } from 'react-router-dom';

// material-ui
import { Box } from '@mui/material';

// project imports
import LogoSection from '../LogoSection';
import { SidebarTrigger } from '@/components/ui/sidebar';

// assets
// import { Icon } from '@iconify/react';

// ==============================|| MAIN NAVBAR / HEADER ||============================== //

const Header = () => {
    const location = useLocation();

    // 检查当前路径是否为面板/控制台页面
    const isConsoleRoute = location.pathname.startsWith('/panel');

    return (
        <>
            {/* logo & sidebar trigger */}
            <Box
                sx={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 2
                }}
            >
                <SidebarTrigger className="h-9 w-9" />
                <Box component="span" sx={{ display: { xs: 'none', md: 'block' } }}>
                    <LogoSection />
                </Box>
            </Box>

            <Box sx={{ flexGrow: 1 }} />
        </>
    );
};

Header.propTypes = {};

export default Header;
