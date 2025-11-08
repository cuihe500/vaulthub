import { createStore } from 'vuex'
import { getToken, setToken, removeToken } from '@/utils/storage'

export default createStore({
  state: {
    token: getToken() || '',
    userInfo: null
  },

  mutations: {
    SET_TOKEN(state, token) {
      state.token = token
      setToken(token)
    },
    REMOVE_TOKEN(state) {
      state.token = ''
      removeToken()
    },
    SET_USER_INFO(state, userInfo) {
      state.userInfo = userInfo
    },
    CLEAR_USER_INFO(state) {
      state.userInfo = null
    }
  },

  actions: {
    // 登录
    login({ commit }, token) {
      commit('SET_TOKEN', token)
    },
    // 登出
    logout({ commit }) {
      commit('REMOVE_TOKEN')
      commit('CLEAR_USER_INFO')
    },
    // 设置用户信息
    setUserInfo({ commit }, userInfo) {
      commit('SET_USER_INFO', userInfo)
    }
  },

  getters: {
    token: state => state.token,
    userInfo: state => state.userInfo,
    isLoggedIn: state => !!state.token
  }
})
