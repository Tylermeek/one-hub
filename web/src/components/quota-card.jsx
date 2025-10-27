'use client';

import * as React from 'react';
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Plus, Info } from 'lucide-react';

import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Progress } from '@/components/ui/progress';
import { useSidebar } from '@/components/ui/sidebar';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { calculateQuota } from '@/utils/common';
import useContext from '@/hooks/useContext';

export function QuotaCard() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { state } = useSidebar();
  const { user } = useSelector((state) => state.account);
  const { currentContext, isTeamContext, getContextQuota } = useContext();

  const [contextQuota, setContextQuota] = React.useState(null);

  // 获取上下文额度信息
  React.useEffect(() => {
    const fetchContextQuota = async () => {
      if (user && currentContext) {
        try {
          const quota = await getContextQuota();
          if (quota) {
            setContextQuota(quota);
          }
        } catch (error) {
          console.error('获取额度信息失败:', error);
        }
      }
    };
    fetchContextQuota();
  }, [user, currentContext, getContextQuota]);

  const handleTopup = () => {
    navigate('/panel/topup');
  };

  // 只在展开状态显示
  if (state === 'collapsed') {
    return null;
  }

  if (!user) {
    return null;
  }

  // 计算进度条数据 - 简化逻辑，只使用 contextQuota
  const totalQuota = contextQuota
    ? contextQuota.unlimited
      ? contextQuota.owner_balance || 0 // 无限额度时使用实际最大额度（所有者余额）
      : contextQuota.available + contextQuota.used_quota
    : user?.quota + (user?.used_quota || 0);

  const usedQuota = contextQuota ? contextQuota.used_quota : user?.used_quota || 0;

  // 计算使用百分比
  const usagePercentage = totalQuota > 0 ? (usedQuota / totalQuota) * 100 : 0;

  // 只有团队上下文才可能无限额度
  const isUnlimited = isTeamContext && contextQuota?.unlimited;
  const displayTotal = isUnlimited ? '∞' : '$' + calculateQuota(totalQuota);
  const displayUsed = '$' + calculateQuota(usedQuota);

  // 获取实际最大额度用于 tooltip 显示
  const actualMaxQuota = isUnlimited ? contextQuota?.owner_balance || 0 : totalQuota;

  return (
    <TooltipProvider>
      <Card className="mx-2 mb-2">
        <CardContent className="p-4">
          <div className="space-y-3">
            {/* 标题和总额度 */}
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium">{isTeamContext ? '团队额度' : '个人额度'}</span>
              <div className="flex items-center gap-1">
                <span className="text-sm text-muted-foreground">{displayTotal}</span>
                {isUnlimited && (
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Info className="h-3 w-3 text-muted-foreground cursor-help" />
                    </TooltipTrigger>
                    <TooltipContent>
                      <p className="text-xs">
                        无限额度模式，实际可用额度为 {isTeamContext ? '团队所有者' : '您'} 的个人余额：$
                        {calculateQuota(actualMaxQuota)}
                      </p>
                    </TooltipContent>
                  </Tooltip>
                )}
              </div>
            </div>

            {/* 进度条 */}
            <div className="space-y-1">
              <div className="flex items-center justify-between text-xs">
                <span className="text-muted-foreground">已使用 {displayUsed}</span>
                <span className="text-muted-foreground">{`${usagePercentage.toFixed(1)}%`}</span>
              </div>
              <Progress value={usagePercentage} className="h-2" max={100} />
            </div>

            {/* 充值按钮 */}
            <Button onClick={handleTopup} className="w-full" size="sm">
              <Plus className="h-4 w-4 mr-2" />
              {t('topup')}
            </Button>
          </div>
        </CardContent>
      </Card>
    </TooltipProvider>
  );
}
