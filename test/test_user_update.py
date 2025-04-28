# -*- coding: utf-8 -*-
"""
test_user_update.py
更新用户信息接口（/api/user/update）API单元测试

- 覆盖正常更新、未登录、参数缺失、token异常等情况
- 需先启动后端服务
- 依赖 pytest fixture auto_register_and_login
"""
import requests
import pytest

BASE_URL = "http://127.0.0.1:8888"
UPDATE_URL = f"{BASE_URL}/api/user/update"

class TestUserUpdate:
    """
    /api/user/update 更新用户信息接口测试
    """
    def test_update_success(self, auto_register_and_login):
        """
        正常更新用户昵称
        """
        token = auto_register_and_login
        headers = {"Authorization": f"Bearer {token}"}
        payload = {"nickname": "新昵称test"}
        resp = requests.put(UPDATE_URL, json=payload, headers=headers)
        assert resp.status_code == 200
        data = resp.json()
        assert data["code"] == 200
        assert data["message"] == "更新成功"

    def test_update_unauthorized(self):
        """
        未携带token，鉴权失败
        """
        payload = {"nickname": "未授权"}
        resp = requests.put(UPDATE_URL, json=payload)
        assert resp.status_code in (401, 403)

    def test_update_missing_params(self, auto_register_and_login):
        """
        缺少参数，接口应返回错误
        """
        token = auto_register_and_login
        headers = {"Authorization": f"Bearer {token}"}
        payload = {}  # 不传任何参数
        resp = requests.put(UPDATE_URL, json=payload, headers=headers)
        assert resp.status_code in (400, 200)
        data = resp.json()
        assert data["code"] in (400, 500)

    def test_update_invalid_token(self):
        """
        携带无效token，鉴权失败
        """
        headers = {"Authorization": "Bearer invalidtoken123"}
        payload = {"nickname": "无效token"}
        resp = requests.put(UPDATE_URL, json=payload, headers=headers)
        assert resp.status_code in (401, 403)

# 运行方法：pytest test_user_update.py
