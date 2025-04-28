# -*- coding: utf-8 -*-
"""
test_user_info.py
获取用户信息接口（/api/user/info）API单元测试

- 覆盖正常获取、未登录、token异常等情况
- 需先启动后端服务
- 依赖 pytest fixture auto_register_and_login
"""
import requests
import pytest

BASE_URL = "http://127.0.0.1:8888"
INFO_URL = f"{BASE_URL}/api/user/info"

class TestUserInfo:
    """
    /api/user/info 获取用户信息接口测试
    """
    def test_get_info_success(self, auto_register_and_login):
        """
        正常获取用户信息
        """
        token = auto_register_and_login
        headers = {"Authorization": f"Bearer {token}"}
        resp = requests.get(INFO_URL, headers=headers)
        assert resp.status_code == 200
        data = resp.json()
        assert data["code"] == 200
        assert data["message"] == "获取成功"
        assert "user" in data

    def test_get_info_unauthorized(self):
        """
        未携带token，鉴权失败
        """
        resp = requests.get(INFO_URL)
        assert resp.status_code in (401, 403)

    def test_get_info_invalid_token(self):
        """
        携带无效token，鉴权失败
        """
        headers = {"Authorization": "Bearer invalidtoken123"}
        resp = requests.get(INFO_URL, headers=headers)
        assert resp.status_code in (401, 403)

# 运行方法：pytest test_user_info.py
