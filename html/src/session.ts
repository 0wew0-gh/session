interface msgItem {
  id: string;
  sid: string;
  content: string;
  time: number;
  userID: string;
}
interface MessageItem {
  msg: msgItem[];
  user: any;
}
interface sendMsgItem {
  mail: string;
  sessionID: string;
  userID: string;
}

export default class Session {
  api = "http://192.168.1.46:36040";
  api2 = "http://192.168.1.46:36041";

  openSessionView: HTMLButtonElement = document.getElementById(
    "open-session-view"
  ) as HTMLButtonElement;

  messageList: msgItem[] = []; // 消息列表
  userMap: any | null = null; // 用户列表
  selfMail: string | null = null; // 用户邮箱
  selfID: string | null = null; // 用户ID
  sessionID: string | null = null; // 会话ID
  token: string | null = null;
  source: EventSource | null = null; // SSE对象
  isScroll = true; // 是否滚动到最下
  isSending = false; // 是否正在发送消息

  constructor() {
    this.openSessionView.addEventListener("click", () => {
      this.token = localStorage.getItem("token");
      this.selfMail = localStorage.getItem("user");
      this.sessionID = localStorage.getItem("sessionID");
      if (this.token == null) {
        this.getGuestToken()
          .then(() => {
            console.log(">>> getGuestToken Done <<<");
            this.getMessage();
          })
          .catch((error) => {
            console.log(error);
            // TODO: 提示 获取 token 失败
            return;
          });
      } else {
        this.checkToken()?.then((res) => {
          if (res) {
            console.log(">>> checkToken Done <<<");
            this.getMessage();
          } else {
            this.getGuestToken()
              .then(() => {
                console.log(">>> getGuestToken Done <<<");
                this.getMessage();
              })
              .catch((error) => {
                console.log(error);
                // TODO: 提示 获取 token 失败
                return;
              });
          }
        });
      }
    });
  }

  getGuestToken(): Promise<void> {
    this.token = null;
    this.selfMail = null;
    this.sessionID = null;
    localStorage.removeItem("token");
    localStorage.removeItem("user");
    localStorage.removeItem("sessionID");
    const formData: FormData = new FormData();
    formData.append("g", "1");

    return fetch(this.api + "/login/", {
      method: "POST",
      body: formData,
    })
      .then((response) => response.json())
      .then((res) => {
        // console.log(">>>", res);
        if (res.code == 10000) {
          const data = res.data;
          this.token = data.token;
          this.selfMail = data.mail;
          localStorage.setItem("token", this.token ?? "");
          localStorage.setItem("user", this.selfMail ?? "");
          // console.log("> token >", token);
          // console.log("> mail >", selfMail);
        }
      })
      .catch((error) => {
        console.log(error);
      });
  }

  checkToken(): Promise<boolean> | null {
    if (this.token == null) {
      // TODO: 提示 token 为空
      return null;
    }
    const formData = new FormData();
    formData.append("t", this.token);

    return fetch(`${this.api}/checkToken/`, {
      method: "POST",
      body: formData,
    })
      .then((response) => response.json())
      .then((res) => {
        switch (res.code) {
          case 10000:
            return true;
            break;
          default:
            return false;
            break;
        }
      });
  }

  getMessage() {
    if (this.sessionID == null) {
      return;
    }
    const formData = new FormData();
    formData.append("sid", this.sessionID);
    formData.append("o", "time");
    formData.append("osc", "DESC");
    formData.append("lt", "0,200");
    formData.append("l", "2");

    fetch(this.api + "/getMessage/", {
      method: "POST",
      body: formData,
    })
      .then((response) => response.json())
      .then((res) => {
        let data;
        let user;
        // let mail;
        switch (res.code) {
          case 10000:
            data = res.data.msg;
            user = res.data.user;
            // mail = res.data.mail;
            if (this.messageList.length == 0) {
              this.messageList = data;
            } else {
              this.messageList = this.messageList.concat(data);
            }
            if (this.userMap == null) {
              this.userMap = user;
            } else {
              for (const key in user) {
                if (!Object.prototype.hasOwnProperty.call(this.userMap, key)) {
                  this.userMap[key] = user[key];
                }
              }
            }
            this.handleMessage(this.messageList);
            setTimeout(() => {
              this.scrollBottom(300);
            }, 350);
            break;
          case 10001:
          case 10002:
          case 10003:
            break;
          case 3900:
            this.token = null;
            localStorage.removeItem("token");
            this.getGuestToken()
              .then(() => {
                console.log(">>> getGuestToken Done <<<");
                this.getMessage();
              })
              .catch((error) => {
                console.log(error);
                // TODO: 提示 获取 token 失败
                return;
              });
        }

        this.setJustifyContent();
      })
      .catch((error) => {
        console.log(error);
      });
  }
  sse() {
    this.sseClose();
    if (this.sessionID == null) {
      return;
    }
    const params = new URLSearchParams({
      sid: this.sessionID,
    });
    this.source = new EventSource(`${this.api2}/sse/?${params}`);
    this.source.onmessage = (event) => {
      // console.log(">>> onmessage <<<");
      // console.log(event.data);
      let ed: MessageItem = {
        msg: [],
        user: {},
      };
      const newMsgList: msgItem[] = [];
      try {
        ed = JSON.parse(event.data);
      } catch (error) {
        console.log(error);
        return;
      }
      console.log(">> ed <<", ed);
      const msgs = ed.msg;
      const users = ed.user;
      if (this.userMap == null) {
        this.userMap = users;
      } else {
        this.userMap = Object.assign(this.userMap, users);
      }
      for (let i = 0; i < msgs.length; i++) {
        const e = msgs[i];
        // console.log(e);
        const userID = e.userID;
        newMsgList.push({
          id: "-1",
          sid: this.sessionID ?? "",
          content: e.content,
          time: e.time,
          userID: userID,
        });
        // console.log(">>", newMsgList);
      }
      this.messageList = this.messageList.concat(newMsgList);
      // console.log("=== sse ===");
      // console.log(newMsgList);
      // console.log(messageList);
      // console.log(userMap);
      this.handleMessage(newMsgList);
      this.scrollBottom(300);
    };
  }
  sendMessage() {
    if (this.isSending) {
      return;
    }
    const sendBtn = document.getElementById(
      "session-send-msg-btn"
    ) as HTMLButtonElement;
    // 保存原来的按钮内容
    const originalContent = "<i class=\"mdi mdi-send\"></i>";

    const formData = new FormData();
    if (this.token != null && this.token.length > 0) {
      formData.append("t", this.token);
    } else if (this.selfMail != null && this.selfMail.length > 0) {
      formData.append("m", this.selfMail);
    } else {
      return;
    }
    let isSSE = false;
    if (this.sessionID != null && this.sessionID.length > 0) {
      formData.append("sid", this.sessionID);
    } else {
      isSSE = true;
    }
    const msgInput: HTMLInputElement | null = document.getElementById(
      "session-send-msg-input"
    ) as HTMLInputElement | null;
    if (msgInput == null || msgInput.value.length == 0) {
      // TODO: 提示消息不能为空
      sendBtn.innerHTML = "<i class=\"mdi mdi-alert error\"></i>";
      setTimeout(function () {
        sendBtn.innerHTML = originalContent;
      }, 1000);
      return;
    }
    formData.append("msg", msgInput.value);
    this.isSending = true;
    sendBtn.disabled = true;
    // 将按钮的内容替换为一个加载动画
    sendBtn.innerHTML = "<i class=\"mdi mdi-loading mdi-spin\"></i>";

    fetch(this.api2 + "/sendMsg/", {
      method: "POST",
      body: formData,
    })
      .then((response) => response.json())
      .then((res) => {
        // console.log(">>>", res);
        sendBtn.disabled = false;
        sendBtn.innerHTML = originalContent;
        this.isSending = false;
        switch (res.code) {
          case 10000:
            {
              msgInput.value = "";
              msgInput.style.height = "20px";
              const data: sendMsgItem = res.data;

              this.selfID = data.userID;
              this.sessionID = data.sessionID;
              localStorage.setItem("sessionID", this.sessionID);
              if (isSSE) {
                this.sse();
              }
              if (this.selfMail != data.mail) {
                // TODO: 提示 本地保存的邮箱 与 token 中的邮箱不匹配
                // 考虑是否需要更新 本地保存的邮箱
                // selfMail = data.mail;
                // localStorage.setItem("user", selfMail);
              }
            }
            break;
          case 10001:
          case 10002:
          case 10003:
            break;
          case 3900:
            this.getGuestToken()
              .then(() => {
                console.log(">>> getGuestToken Done <<<");
                this.getMessage();
              })
              .catch((error) => {
                console.log(error);
                // TODO: 提示 获取 token 失败
                return;
              });
            break;
          default:
            console.log("other error", res);
            break;
        }
      })
      .catch((error) => {
        sendBtn.disabled = false;
        sendBtn.innerHTML = originalContent;
        this.isSending = false;
        console.log(error);
      });
  }

  sseClose() {
    if (this.source != null) {
      this.source.close();
      this.source = null;
    }
  }

  handleMessage(msgs: msgItem[]) {
    const sessionContent = document.getElementById(
      "session-content"
    ) as HTMLDivElement | null;
    if (sessionContent == null) {
      return;
    }
    for (let i = 0; i < msgs.length; i++) {
      const msg = msgs[i]; //{content: "112",id: "14",sessionID: "11",time: 1695014999000,updateTime: "",userID: "1"} userID:1 为匿名用户，自己
      // const user = this.userMap[msg.userID];
      // console.log(JSON.stringify(msg));
      // console.log(JSON.stringify(user));
      // console.log(checkUserMail(msg.userID));
      const isSelf = this.checkUserMail(msg.userID);

      // 气泡
      const msgdiv = document.createElement("div");
      msgdiv.className = "msg";
      if (isSelf) {
        msgdiv.classList.add("msg-r");
      } else {
        msgdiv.classList.add("msg-l");
      }

      // 气泡中内容和时间
      const divTime = document.createElement("div");
      divTime.className = "msg-time";
      divTime.innerText = this.getTime(msg.time);
      const divContent = document.createElement("div");
      divContent.innerText = msg.content;
      msgdiv.appendChild(divContent);
      msgdiv.appendChild(divTime);

      // 头像
      const userAvatar = document.createElement("div");
      userAvatar.className = "userAvatar";
      if (isSelf) {
        userAvatar.innerHTML = "<i class=\"mdi mdi-account-circle\"></i>";
      } else {
        userAvatar.innerHTML = "<i class=\"mdi mdi-account-circle-outline\"></i>";
      }

      // 消息
      const div = document.createElement("div");
      div.classList.add("msgview");
      if (isSelf) {
        div.classList.add("msgview-r");
        div.appendChild(msgdiv);
        div.appendChild(userAvatar);
      } else {
        div.appendChild(userAvatar);
        div.appendChild(msgdiv);
      }

      sessionContent.appendChild(div);
    }
    this.setJustifyContent(sessionContent);
  }
  setJustifyContent(sessionContent: HTMLDivElement | null = null) {
    if (sessionContent == null) {
      sessionContent = document.getElementById(
        "session-content"
      ) as HTMLDivElement | null;
    }
    if (sessionContent == null) {
      return;
    }
    let contentHeight = 0;
    const contents = sessionContent.getElementsByClassName("msgview");
    for (let i = 0; i < contents.length; i++) {
      const e = contents[i];
      contentHeight += e.clientHeight;
    }
    if (contentHeight <= sessionContent.clientHeight) {
      sessionContent.style.justifyContent = "flex-end";
    } else if (sessionContent.style.justifyContent != "") {
      sessionContent.style.justifyContent = "";
    }
    console.log(contentHeight, sessionContent.clientHeight);
  }
  getTime(ts: string | number) {
    const date = new Date(ts);
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, "0");
    const day = String(date.getDate()).padStart(2, "0");
    const hour = String(date.getHours()).padStart(2, "0");
    const min = String(date.getMinutes()).padStart(2, "0");
    const sec = String(date.getSeconds()).padStart(2, "0");
    return `${year}-${month}-${day} ${hour}:${min}:${sec}`;
  }
  checkUserMail(id: string) {
    if (id == "-1") {
      return false;
    } else if (id == "-2") {
      return true;
    }

    if (this.userMap == null) {
      return false;
    }
    if (this.userMap[id] == null) {
      // console.log(userMap, id);
      return false;
    }
    const uMail = this.userMap[id].mail;
    if (uMail == this.selfMail) {
      return true;
    }
    return false;
  }
  scrollBottom(duration = 350) {
    if (!this.isScroll) {
      console.log(">>> 禁止滚动到最下 <<<");
      return;
    }
    let start = 0;
    let end = 0;
    let change = 0;
    let startTime: number | null = null;
    if (duration < 300) {
      duration = 300;
    }
    function animateScroll(currentTime: number) {
      if (startTime === null) startTime = currentTime;
      const timeElapsed = currentTime - startTime;
      const progress = Math.min(timeElapsed / duration, 1);
      if (list != null) {
        list.scrollTop = start + change * progress;
      }
      if (progress < 1) {
        window.requestAnimationFrame(animateScroll);
      }
    }
    const list = document.getElementById("session-content");
    if (list != null) {
      start = list.scrollTop;
      end = list.scrollHeight;
      change = end - start;

      if (list != null) {
        const start = list.scrollTop;
        const end = list.scrollHeight;
        change = end - start;
        window.requestAnimationFrame(animateScroll);
      }
    }
  }
}
