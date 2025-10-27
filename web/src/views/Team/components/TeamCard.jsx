import PropTypes from 'prop-types';
import { useNavigate } from 'react-router-dom';
import { Users, TrendingUp, Clock } from 'lucide-react';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { cn } from '@/lib/utils';
import { renderQuota } from 'utils/common';
import { getRoleText, getRoleBadgeVariant } from 'utils/teamPermissions';

const TeamCard = ({ team, className }) => {
  const navigate = useNavigate();

  // 格式化时间显示
  const formatLastActivity = (timestamp) => {
    if (!timestamp) return '无活动';

    const now = Date.now();
    const time = timestamp * 1000;
    const diff = now - time;

    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(diff / 3600000);
    const days = Math.floor(diff / 86400000);

    if (minutes < 60) return `${minutes}分钟前`;
    if (hours < 24) return `${hours}小时前`;
    if (days < 30) return `${days}天前`;
    return new Date(time).toLocaleDateString();
  };

  // 计算额度使用率
  const quotaUsageRate = team.unlimited_quota ? 0 : team.quota > 0 ? (team.used_quota / team.quota) * 100 : 0;

  // 格式化额度显示
  const formatQuota = (quota, unlimited) => {
    if (unlimited) return '∞';
    return renderQuota(quota || 0, 2);
  };

  const handleClick = () => {
    navigate(`/panel/team/${team.id}`);
  };

  return (
    <Card className={cn('cursor-pointer transition-all duration-200 hover:shadow-md', className)} onClick={handleClick}>
      <CardContent className="p-6">
        {/* 头部：团队名称和角色 */}
        <div className="flex items-start justify-between mb-4">
          <div className="flex-1 min-w-0">
            <h3 className="text-lg font-semibold text-foreground truncate mb-1">{team.name}</h3>
            {team.description && <p className="text-sm text-muted-foreground line-clamp-2">{team.description}</p>}
          </div>
          <div className="flex gap-1 ml-2">
            <Badge variant={getRoleBadgeVariant(team.current_user_role)}>{getRoleText(team.current_user_role)}</Badge>
            {team.status === 1 ? (
              <Badge variant="outline" className="text-green-600 border-green-600">
                正常
              </Badge>
            ) : (
              <Badge variant="outline" className="text-red-600 border-red-600">
                禁用
              </Badge>
            )}
          </div>
        </div>

        {/* 额度信息 */}
        <div className="space-y-2 mb-4">
          <div className="flex items-center justify-between text-sm">
            <span className="text-muted-foreground">团队额度</span>
            <span className="font-semibold text-foreground">{formatQuota(team.quota, team.unlimited_quota)}</span>
          </div>
          <div className="flex items-center justify-between text-sm">
            <span className="text-muted-foreground">已用额度</span>
            <span className="font-medium text-foreground">{renderQuota(team.used_quota || 0, 2)}</span>
          </div>
          {!team.unlimited_quota && team.quota > 0 && (
            <div className="space-y-1">
              <Progress value={quotaUsageRate} className="h-2" />
              <div className="flex items-center justify-between text-xs text-muted-foreground">
                <span>使用率</span>
                <span>{quotaUsageRate.toFixed(1)}%</span>
              </div>
            </div>
          )}
        </div>

        {/* 底部信息 */}
        <div className="flex items-center justify-between pt-3 border-t border-border">
          <div className="flex items-center gap-4 text-sm text-muted-foreground">
            <div className="flex items-center gap-1">
              <Users className="w-4 h-4" />
              <span>{team.member_count || 0} 成员</span>
            </div>
            {team.request_count_7d !== undefined && (
              <div className="flex items-center gap-1">
                <TrendingUp className="w-4 h-4" />
                <span>{team.request_count_7d} 请求</span>
              </div>
            )}
          </div>
          <div className="flex items-center gap-1 text-xs text-muted-foreground">
            <Clock className="w-3 h-3" />
            <span>{formatLastActivity(team.last_activity_time)}</span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};

TeamCard.propTypes = {
  team: PropTypes.shape({
    id: PropTypes.number.isRequired,
    name: PropTypes.string.isRequired,
    description: PropTypes.string,
    quota: PropTypes.number,
    used_quota: PropTypes.number,
    unlimited_quota: PropTypes.bool,
    status: PropTypes.number,
    current_user_role: PropTypes.number,
    is_owner: PropTypes.bool,
    member_count: PropTypes.number,
    request_count_7d: PropTypes.number,
    last_activity_time: PropTypes.number
  }).isRequired,
  className: PropTypes.string
};

export default TeamCard;
