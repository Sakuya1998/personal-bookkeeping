import React, { useEffect, useState } from 'react';
import { Modal, Table, Button, Select, Input, message, Popconfirm, Space, Tag, App } from 'antd';
import { PlusOutlined, DeleteOutlined, UserSwitchOutlined, LogoutOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import client from '../api/client';
import { ApiResponse, LedgerMember, MemberRole } from '../api/types';
import { useAppStore } from '../store/appStore';

interface Props {
  open: boolean;
  ledgerId: string;
  onClose: () => void;
}

const MembersModal: React.FC<Props> = ({ open, ledgerId, onClose }) => {
  const { t } = useTranslation();
  const { user, currentRole } = useAppStore();
  const [members, setMembers] = useState<LedgerMember[]>([]);
  const [loading, setLoading] = useState(false);
  const [inviteUsername, setInviteUsername] = useState('');
  const [inviting, setInviting] = useState(false);

  const isOwner = currentRole === 'owner';
  const canManage = currentRole === 'owner' || currentRole === 'admin';

  const loadMembers = async () => {
    setLoading(true);
    try {
      const res = await client.get<ApiResponse<LedgerMember[]>>(`/ledgers/${ledgerId}/members`);
      setMembers(res.data.data);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (open && ledgerId) loadMembers();
  }, [open, ledgerId]);

  const handleInvite = async () => {
    if (!inviteUsername.trim()) return;
    setInviting(true);
    try {
      await client.post(`/ledgers/${ledgerId}/members`, { username: inviteUsername.trim() });
      message.success(t('members.inviteSuccess'));
      setInviteUsername('');
      loadMembers();
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || t('common.failed'));
    } finally {
      setInviting(false);
    }
  };

  const handleRemove = async (targetUserId: string) => {
    try {
      await client.delete(`/ledgers/${ledgerId}/members/${targetUserId}`);
      message.success(t('members.removeSuccess'));
      loadMembers();
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || t('common.failed'));
    }
  };

  const handleLeave = async () => {
    try {
      await client.post(`/ledgers/${ledgerId}/leave`);
      message.success(t('members.leaveSuccess'));
      onClose();
      // Refresh ledgers list in parent
      window.location.reload();
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || t('common.failed'));
    }
  };

  const handleUpdateRole = async (targetUserId: string, role: MemberRole) => {
    try {
      await client.put(`/ledgers/${ledgerId}/members/${targetUserId}`, { role });
      message.success(t('members.roleUpdateSuccess'));
      loadMembers();
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || t('common.failed'));
    }
  };

  const roleColors: Record<string, string> = {
    owner: 'gold',
    admin: 'blue',
    member: 'default',
  };

  const columns = [
    { title: t('members.username'), dataIndex: 'username', key: 'username' },
    {
      title: t('members.role'), dataIndex: 'role', key: 'role', width: 150,
      render: (role: MemberRole, record: LedgerMember) => {
        if (isOwner && record.user_id !== user?.id) {
          return (
            <Select
              size="small"
              value={role}
              style={{ width: 110 }}
              onChange={(v) => handleUpdateRole(record.user_id, v)}
              options={[
                { label: t('members.admin'), value: 'admin' },
                { label: t('members.member'), value: 'member' },
              ]}
            />
          );
        }
        return <Tag color={roleColors[role]}>{t(`members.${role}`)}</Tag>;
      },
    },
    {
      title: t('members.joinedAt'), dataIndex: 'joined_at', key: 'joined_at', width: 120,
      render: (v: string) => v ? v.slice(0, 10) : '-',
    },
    {
      title: t('members.action'), key: 'action', width: 80,
      render: (_: unknown, record: LedgerMember) => {
        if (record.user_id === user?.id) return null;
        if (!canManage) return null;
        return (
          <Popconfirm
            title={t('members.removeConfirm', { username: record.username })}
            onConfirm={() => handleRemove(record.user_id)}
          >
            <Button size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        );
      },
    },
  ];

  return (
    <Modal
      title={t('members.title')}
      open={open}
      onCancel={onClose}
      footer={null}
      width={600}
    >
      {canManage && (
        <Space.Compact style={{ width: '100%', marginBottom: 16 }}>
          <Input
            placeholder={t('members.invitePlaceholder')}
            value={inviteUsername}
            onChange={(e) => setInviteUsername(e.target.value)}
            onPressEnter={handleInvite}
          />
          <Button type="primary" icon={<PlusOutlined />} loading={inviting} onClick={handleInvite}>
            {t('members.invite')}
          </Button>
        </Space.Compact>
      )}

      <Table
        dataSource={members}
        columns={columns}
        rowKey="id"
        loading={loading}
        size="small"
        pagination={false}
      />

      <div style={{ marginTop: 16 }}>
        {!isOwner && (
          <Popconfirm title={t('members.leaveConfirm')} onConfirm={handleLeave}>
            <Button danger icon={<LogoutOutlined />}>{t('members.leave')}</Button>
          </Popconfirm>
        )}
      </div>
    </Modal>
  );
};

export default MembersModal;
