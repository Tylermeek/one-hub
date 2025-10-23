import { Icon } from '@iconify/react';

const icons = {
    IconUsers: () => <Icon width={20} icon="solar:users-group-rounded-bold-duotone" />,
    IconUserPlus: () => <Icon width={20} icon="solar:user-plus-bold-duotone" />,
    IconUserMinus: () => <Icon width={20} icon="solar:user-minus-bold-duotone" />,
    IconWallet: () => <Icon width={20} icon="solar:wallet-money-bold-duotone" />,
    IconChart: () => <Icon width={20} icon="solar:chart-2-bold-duotone" />
};

const team = {
    id: 'team',
    title: '团队管理',
    type: 'group',
    children: [
        {
            id: 'team',
            title: '我的团队',
            type: 'item',
            url: '/panel/team',
            icon: icons.IconUsers,
            breadcrumbs: false,
            isAdmin: false
        }
    ]
};

export default team;

