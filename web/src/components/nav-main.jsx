import { ChevronRight } from 'lucide-react';
import { Link, useLocation } from 'react-router-dom';
import { useSelector } from 'react-redux';
import { useTranslation } from 'react-i18next';

import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import {
    SidebarGroup,
    SidebarGroupLabel,
    SidebarMenu,
    SidebarMenuButton,
    SidebarMenuItem,
    SidebarMenuSub,
    SidebarMenuSubButton,
    SidebarMenuSubItem
} from '@/components/ui/sidebar';
import { useIsAdmin } from '@/utils/common';
import menuItem from '@/menu-items';

export function NavMain() {
    const { t } = useTranslation();
    const location = useLocation();
    const userIsAdmin = useIsAdmin();
    const siteInfo = useSelector((state) => state.siteInfo);

    // 转换菜单数据为 NavMain 所需格式
    const convertMenuItems = (items) => {
        return items
            .map((group) => {
                const filteredChildren = group.children.filter(
                    (child) => (!child.isAdmin || userIsAdmin) && !(siteInfo.UserInvoiceMonth === false && child.id === 'invoice')
                );

                if (filteredChildren.length === 0) {
                    return null;
                }

                return {
                    title: t(group.id),
                    icon: null, // 暂时不显示分组图标
                    isActive: false,
                    items: filteredChildren.map((child) => {
                        if (child.type === 'collapse') {
                            return {
                                title: t(child.id),
                                url: '#',
                                icon: child.icon,
                                children: child.children?.map((subChild) => ({
                                    title: t(subChild.id),
                                    url: subChild.url,
                                    icon: subChild.icon
                                }))
                            };
                        } else {
                            return {
                                title: t(child.id),
                                url: child.url,
                                icon: child.icon
                            };
                        }
                    })
                };
            })
            .filter(Boolean);
    };

    const menuData = convertMenuItems(menuItem.items);

    const isActive = (url) => {
        return location.pathname === url;
    };

    const renderMenuItem = (item, level = 0) => {
        if (item.children && item.children.length > 0) {
            // 嵌套菜单项
            return (
                <Collapsible key={item.title} asChild defaultOpen={item.isActive} className="group/collapsible">
                    <SidebarMenuItem>
                        <CollapsibleTrigger asChild>
                            <SidebarMenuButton tooltip={item.title}>
                                {item.icon && <item.icon />}
                                <span>{item.title}</span>
                                <ChevronRight className="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90" />
                            </SidebarMenuButton>
                        </CollapsibleTrigger>
                        <CollapsibleContent>
                            <SidebarMenuSub>
                                {item.children.map((subItem) => (
                                    <SidebarMenuSubItem key={subItem.title}>
                                        {subItem.children ? (
                                            // 三级菜单
                                            <Collapsible asChild className="group/collapsible">
                                                <SidebarMenuItem>
                                                    <CollapsibleTrigger asChild>
                                                        <SidebarMenuSubButton>
                                                            {subItem.icon && <subItem.icon />}
                                                            <span>{subItem.title}</span>
                                                            <ChevronRight className="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90" />
                                                        </SidebarMenuSubButton>
                                                    </CollapsibleTrigger>
                                                    <CollapsibleContent>
                                                        <SidebarMenuSub>
                                                            {subItem.children.map((subSubItem) => (
                                                                <SidebarMenuSubItem key={subSubItem.title}>
                                                                    <SidebarMenuSubButton asChild>
                                                                        <Link
                                                                            to={subSubItem.url}
                                                                            className={isActive(subSubItem.url) ? 'bg-sidebar-accent' : ''}
                                                                        >
                                                                            {subSubItem.icon && <subSubItem.icon />}
                                                                            <span>{subSubItem.title}</span>
                                                                        </Link>
                                                                    </SidebarMenuSubButton>
                                                                </SidebarMenuSubItem>
                                                            ))}
                                                        </SidebarMenuSub>
                                                    </CollapsibleContent>
                                                </SidebarMenuItem>
                                            </Collapsible>
                                        ) : (
                                            // 二级菜单
                                            <SidebarMenuSubButton asChild>
                                                <Link to={subItem.url} className={isActive(subItem.url) ? 'bg-sidebar-accent' : ''}>
                                                    {subItem.icon && <subItem.icon />}
                                                    <span>{subItem.title}</span>
                                                </Link>
                                            </SidebarMenuSubButton>
                                        )}
                                    </SidebarMenuSubItem>
                                ))}
                            </SidebarMenuSub>
                        </CollapsibleContent>
                    </SidebarMenuItem>
                </Collapsible>
            );
        } else {
            // 普通菜单项
            return (
                <SidebarMenuItem key={item.title}>
                    <SidebarMenuButton asChild tooltip={item.title}>
                        <Link to={item.url} className={isActive(item.url) ? 'bg-sidebar-accent' : ''}>
                            {item.icon && <item.icon />}
                            <span>{item.title}</span>
                        </Link>
                    </SidebarMenuButton>
                </SidebarMenuItem>
            );
        }
    };

    return (
        <>
            {menuData.map((group) => (
                <SidebarGroup key={group.title}>
                    <SidebarGroupLabel>{group.title}</SidebarGroupLabel>
                    <SidebarMenu>{group.items.map((item) => renderMenuItem(item))}</SidebarMenu>
                </SidebarGroup>
            ))}
        </>
    );
}
