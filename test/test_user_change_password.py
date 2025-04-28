# -*- coding: utf-8 -*-
"""
test_user_change_password.py
修改用户密码接口（/api/user/change_password）API单元测试

- 覆盖正常修改、旧密码错误、未登录、参数缺失、token异常等情况
- 需先启动后端服务
- 依赖 pytest fixture auto_register_and_login
"""
import requests
import pytest
from conftest import TEST_PASSWORD

BASE_URL = "http://127.0.0.1:8888"
CHANGE_PWD_URL = f"{BASE_URL}/api/user/change_password"

class TestUserChangePassword:
    """
    /api/user/change_password 修改密码接口测试
    """
    def test_change_password_success(self, auto_register_and_login):
        """
        正常修改密码（改回原密码，保证幂等）
        """
        token = auto_register_and_login
        headers = {"Authorization": f"Bearer {token}"}
        payload = {"old_password": TEST_PASSWORD, "new_password": TEST_PASSWORD}
        resp = requests.post(CHANGE_PWD_URL, json=payload, headers=headers)
        assert resp.status_code == 200
        data = resp.json()
        assert data["code"] == 200
        assert "成功" in data["message"]

    def test_change_password_wrong_old(self, auto_register_and_login):
        """
        旧密码错误
        """
        token = auto_register_and_login
        headers = {"Authorization": f"Bearer {token}"}
        payload = {"old_password": "wrongpass", "new_password": TEST_PASSWORD}
        resp = requests.post(CHANGE_PWD_URL, json=payload, headers=headers)
        assert resp.status_code == 200
        data = resp.json()
        assert data["code"] != 200
        assert "错误" in data["message"] or "失败" in data["message"]

    def test_change_password_unauthorized(self):
        """
        未携带token，鉴权失败
        """
        payload = {"old_password": "any", "new_password": "any"}
        resp = requests.post(CHANGE_PWD_URL, json=payload)
        assert resp.status_code in (401, 403)

    def test_change_password_missing_params(self, auto_register_and_login):
        """
        缺少参数
        """
        token = auto_register_and_login
        headers = {"Authorization": f"Bearer {token}"}
        payload = {"old_password": TEST_PASSWORD}  # 缺 new_password
        resp = requests.post(CHANGE_PWD_URL, json=payload, headers=headers)
        assert resp.status_code in (400, 200)
        data = resp.json()
        assert data["code"] in (400, 500)

    def test_change_password_invalid_token(self):
        """
        携带无效token，鉴权失败
        """
        headers = {"Authorization": "Bearer invalidtoken123"}
        payload = {"old_password": "any", "new_password": "any"}
        resp = requests.post(CHANGE_PWD_URL, json=payload, headers=headers)
        assert resp.status_code in (401, 403)

# 运行方法：pytest test_user_change_password.py
