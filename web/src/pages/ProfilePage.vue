<script setup lang="ts">
import { api } from '../services/api'
const userID = localStorage.getItem('demo_user') ?? 'u-volunteer'
</script>

<template><section class="profile"><div class="avatar">周</div><div><span class="eyebrow">社区志愿者</span><h2>{{ userID === 'u-captain' ? '林晓队长' : '周宁' }}</h2><p>已确认服务记录会形成个人档案；退出小队不会删除历史记录。</p><button class="primary" @click="api.downloadCertificate(userID)">导出服务证明</button></div></section></template>
