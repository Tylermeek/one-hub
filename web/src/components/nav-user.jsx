import { useNavigate } from 'react-router-dom';
import { useSelector, useDispatch } from 'react-redux';
import { ChevronsUpDown, CreditCard, LogOut, Settings, User, Bell, Sun, Moon, Languages } from 'lucide-react';

import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuGroup,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuSub,
    DropdownMenuSubContent,
    DropdownMenuSubTrigger,
    DropdownMenuTrigger
} from '@/components/ui/dropdown-menu';
import { SidebarMenu, SidebarMenuButton, SidebarMenuItem, useSidebar } from '@/components/ui/sidebar';
import { LOGOUT, SET_THEME } from '@/store/actions';
import { useNotice } from '@/ui-component/notice/NoticeContext';
import useI18n from '@/hooks/useI18n';
import i18nList from '@/i18n/i18nList';
import Flags from 'country-flag-icons/react/3x2';

export function NavUser() {
    const { isMobile } = useSidebar();
    const navigate = useNavigate();
    const dispatch = useDispatch();
    const { user } = useSelector((state) => state.account);
    const defaultTheme = useSelector((state) => state.customization.theme);
    const { openNotice } = useNotice();
    const i18n = useI18n();

    const handleLogout = () => {
        dispatch({ type: LOGOUT });
        navigate('/login');
    };

    const handleProfile = () => {
        navigate('/panel/profile');
    };

    const handleBilling = () => {
        navigate('/panel/topup');
    };

    const handleSettings = () => {
        navigate('/panel/setting');
    };

    const handleNotice = () => {
        openNotice();
    };

    const handleThemeToggle = () => {
        const newTheme = defaultTheme === 'light' ? 'dark' : 'light';
        dispatch({ type: SET_THEME, theme: newTheme });
        localStorage.setItem('theme', newTheme);
    };

    const handleLanguageChange = (lng) => {
        i18n.changeLanguage(lng);
    };

    if (!user) {
        return null;
    }

    return (
        <SidebarMenu>
            <SidebarMenuItem>
                <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                        <SidebarMenuButton
                            size="lg"
                            className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
                        >
                            <Avatar className="h-8 w-8 rounded-lg">
                                <AvatarImage src={user.avatar_url} alt={user.username} />
                                <AvatarFallback className="rounded-lg">{user.username?.charAt(0)?.toUpperCase() || 'U'}</AvatarFallback>
                            </Avatar>
                            <div className="grid flex-1 text-left text-sm leading-tight">
                                <span className="truncate font-medium">{user.username}</span>
                                <span className="truncate text-xs">{user.email}</span>
                            </div>
                            <ChevronsUpDown className="ml-auto size-4" />
                        </SidebarMenuButton>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent
                        className="w-56 min-w-56 rounded-lg z-50 bg-white border shadow-lg"
                        side={isMobile ? 'bottom' : 'right'}
                        align="end"
                        sideOffset={4}
                    >
                        <DropdownMenuLabel className="p-0 font-normal">
                            <div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
                                <Avatar className="h-8 w-8 rounded-lg">
                                    <AvatarImage src={user.avatar_url} alt={user.username} />
                                    <AvatarFallback className="rounded-lg">{user.username?.charAt(0)?.toUpperCase() || 'U'}</AvatarFallback>
                                </Avatar>
                                <div className="grid flex-1 text-left text-sm leading-tight">
                                    <span className="truncate font-medium">{user.username}</span>
                                    <span className="truncate text-xs">{user.email}</span>
                                </div>
                            </div>
                        </DropdownMenuLabel>
                        <DropdownMenuSeparator />
                        <DropdownMenuGroup>
                            <DropdownMenuItem onClick={handleProfile}>
                                <User />
                                个人设置
                            </DropdownMenuItem>
                            <DropdownMenuItem onClick={handleBilling}>
                                <CreditCard />
                                充值
                            </DropdownMenuItem>
                            <DropdownMenuItem onClick={handleSettings}>
                                <Settings />
                                系统设置
                            </DropdownMenuItem>
                        </DropdownMenuGroup>
                        <DropdownMenuSeparator />
                        <DropdownMenuGroup>
                            <DropdownMenuItem onClick={handleNotice}>
                                <Bell />
                                通知
                            </DropdownMenuItem>
                            <DropdownMenuItem onClick={handleThemeToggle}>
                                {defaultTheme === 'light' ? <Sun /> : <Moon />}
                                主题: {defaultTheme === 'light' ? '浅色' : '深色'}
                            </DropdownMenuItem>
                            <DropdownMenuSub>
                                <DropdownMenuSubTrigger>
                                    <Languages />
                                    语言: {i18nList.find((item) => item.lng === i18n.language)?.name || '简体中文'}
                                </DropdownMenuSubTrigger>
                                <DropdownMenuSubContent>
                                    {i18nList.map((item) => {
                                        const FlagComponent = Flags[item.countryCode];
                                        return (
                                            <DropdownMenuItem
                                                key={item.lng}
                                                onClick={() => handleLanguageChange(item.lng)}
                                                className="gap-2"
                                            >
                                                {FlagComponent && (
                                                    <div className="flex size-4 items-center justify-center">
                                                        <FlagComponent className="size-4" />
                                                    </div>
                                                )}
                                                {item.name}
                                            </DropdownMenuItem>
                                        );
                                    })}
                                </DropdownMenuSubContent>
                            </DropdownMenuSub>
                        </DropdownMenuGroup>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem onClick={handleLogout}>
                            <LogOut />
                            退出登录
                        </DropdownMenuItem>
                    </DropdownMenuContent>
                </DropdownMenu>
            </SidebarMenuItem>
        </SidebarMenu>
    );
}
