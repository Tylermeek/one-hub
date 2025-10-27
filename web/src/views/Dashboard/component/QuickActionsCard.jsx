import React from 'react';
import { Key, FileText, Users, CreditCard, Settings, HelpCircle, ExternalLink } from 'lucide-react';
import { useTranslation } from 'react-i18next';

// shadcn components
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

const QuickActionsCard = ({ className = '' }) => {
  const { t } = useTranslation();

  // 快速操作配置
  const quickActions = [
    {
      id: 'token',
      title: '创建 Token',
      description: '生成新的 API 密钥',
      icon: Key,
      href: '/token',
      color: 'bg-blue-500'
    },
    {
      id: 'logs',
      title: '查看日志',
      description: '查看 API 调用记录',
      icon: FileText,
      href: '/log',
      color: 'bg-green-500'
    },
    {
      id: 'team',
      title: '管理团队',
      description: '创建或管理团队',
      icon: Users,
      href: '/team',
      color: 'bg-purple-500'
    },
    {
      id: 'payment',
      title: '充值额度',
      description: '购买更多额度',
      icon: CreditCard,
      href: '/payment',
      color: 'bg-orange-500'
    },
    {
      id: 'settings',
      title: '个人设置',
      description: '修改个人信息',
      icon: Settings,
      href: '/profile',
      color: 'bg-gray-500'
    },
    {
      id: 'help',
      title: '帮助文档',
      description: '查看使用指南',
      icon: HelpCircle,
      href: '/docs',
      color: 'bg-indigo-500',
      external: true
    }
  ];

  const handleActionClick = (action) => {
    if (action.external) {
      window.open(action.href, '_blank');
    } else {
      window.location.href = action.href;
    }
  };

  return (
    <Card className={cn('h-auto', className)}>
      <CardHeader className="pb-3">
        <CardTitle className="flex items-center gap-2 text-lg">
          <Settings className="w-4 h-4" />
          {t('dashboard_index.quick_actions')}
        </CardTitle>
      </CardHeader>
      <CardContent className="pt-0">
        <div className="grid grid-cols-2 gap-2">
          {quickActions.map((action) => {
            const IconComponent = action.icon;
            return (
              <Button
                key={action.id}
                variant="outline"
                className="h-auto p-3 flex flex-col items-start gap-1.5 hover:bg-muted/50 transition-colors"
                onClick={() => handleActionClick(action)}
              >
                <div className="flex items-center gap-2 w-full">
                  <div className={cn('p-1.5 rounded-lg text-white', action.color)}>
                    <IconComponent className="w-3.5 h-3.5" />
                  </div>
                  <div className="flex-1 text-left">
                    <div className="font-medium text-sm">{action.title}</div>
                  </div>
                </div>
              </Button>
            );
          })}
        </div>

        {/* 底部提示 */}
        <div className="mt-3 p-2 bg-muted/30 rounded-lg">
          <p className="text-xs text-muted-foreground text-center">点击任意操作快速跳转到对应功能页面</p>
        </div>
      </CardContent>
    </Card>
  );
};

export default QuickActionsCard;
