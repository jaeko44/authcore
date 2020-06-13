/**
 * Using per-page grouping for translation key/value pair, the base key corresponds to the page.
 * Common keys for each page:
 *
 * title: Title of the page, shows above of the separator line of the widget
 * description: Description of the page, shows under the separator line of the widget
 * list_item: Exists when the page includes list element, includes title, text or links in the list
 * input: Exists when the page includes input field, includes label and error of the input field
 * button: Exists when the page includes button
 * link: Exists when the page includes link
 * text: Miscellaneous text in the page

 * There is a section for generally used translation at the end.
 **/

export default {
  'zh-HK': {
    password_strength_indicator: {
      title: {
        password_strength: '密碼強度：',
        too_weak: '弱',
        ok: '尚可',
        strong: '強'
      },
      description: {
        too_weak: '你可混合使用數字、大階、小階和符號來強化密碼。',
        ok: '不錯。還可以再強一點。',
        strong: '很好。你的密碼十分安全。'
      }
    },
    register: {
      title: '建立帳戶',
      description: {
        start: '選擇以下方式',
        one_time_code_sent: '一次認證碼傳送到'
      },
      input: {
        label: {
          contact: '電郵或電話號碼',
          username: '用戶名稱',
          password: '密碼',
          one_time_code: '一次認證碼'
        },
        error: {
          duplicate_username: '用戶名稱已有人使用',
          invalid_username: '用戶名稱無效',
          invalid_email: '電郵無效',
          invalid_phone: '電話號碼無效',
          invalid_contact: '聯絡方法無效',
          requires_better_password_strength: '密碼強度不足',
          sign_in: '登入',
          email_already_exists: '電郵已有人使用，請嘗試{link}。',
          phone_already_exists: '電話號碼已有人使用，請嘗試{link}。'
        }
      },
      button: {
        register: '建立帳戶',
        next: '下一步'
      },
      link: {
        privacy_link: '服務條款及隱私權政策',
        sign_in: '登入',
        resend_code: '再新發送認證碼'
      },
      text: {
        // privacy_policy: {link} will be replaced by link to privacy policy
        privacy_policy: '我同意 {link}',
        or: '或'
      }
    },
    add_recovery_email: {
      title: '還原電郵',
      description: '增加還原電郵，確保你有需要時可以重新登入。',
      input: {
        label: {
          recovery_email: '還原電郵'
        }
      },
      button: {
        next: '下一步'
      }
    },
    verification: {
      title: {
        account_verification: '認證帳戶',
        contact_verification: '認證電郵或電話號碼',
        account_created: '成功建立帳戶'
      },
      input: {
        label: {
          verification_code: '六位數字認證碼'
        },
        error: {
          too_frequent: '請稍後再試',
          invalid_verification_code: '認證碼錯誤'
        }
      },
      button: {
        verify: '認證',
        ok: '好'
      },
      link: {
        resend_code: '再新發送認證碼',
        verify_later: '稍候再說',
        cancel: '取消'
      },
      text: {
        // contact_verification: {contact} will be replaced by contact which user inputs
        contact_verification: '六位數字認證碼已發送到 {contact}',
        code_sent: '認證碼已重新發送'
      }
    },
    resend_verification: {
      title: '再新發送認證碼',
      text: {
        // resend_contact_verification: {contact} will be replaced by contact which user inputs
        resend_contact_verification: '六位數字認證碼已發送到 {contact}'
      }
    },
    sign_in: {
      title: {
        sign_in: '登入',
        continue: '登入',
        two_step_verification: '雙重認證',
        password: '密碼'
      },
      description: {
        continue: '選擇以下方式',
        enter_password: '輸入密碼登入',
        two_step_verification: '使用其他方式',
        error: {
          used_contact_in_system: '此聯絡方法已被使用'
        }
      },
      list_item: {
        title: {
          security_key: '安全金鑰',
          sms_code: '簡訊',
          authenticator_app: '認證應用程式',
          backup_code: '備用認證碼'
        },
        text: {
          sms_code: '輸入發送到電話簡訊的六位數字認證碼',
          authenticator_app: '輸入由Authenticator app 產生的六位數字認證碼',
          backup_code: '輸入任何一組八位數字的備用認證碼'
        }
      },
      input: {
        label: {
          handle: '電郵、電話號碼或用戶名稱',
          contact: '電郵或電話號碼',
          password: '密碼',
          sms_code: '六位數字認證碼',
          authenticator_app: '六位數字認證碼',
          backup_code: '八位數字備用認證碼'
        },
        error: {
          blank: '此欄不能留空',
          user_is_locked: '用戶已被鎖定',
          user_not_found_not_allow_create_account: '找不到用戶',
          user_not_found_allow_create_account: '找不到用戶，請嘗試 {link}',
          password_not_set: '你的帳戶以社群登入方式建立，請嘗試其中一種社群登入',
          incorrect_password: '密碼錯誤',
          invalid_totp_pin: '認證碼錯誤',
          invalid_sms_code: '認證碼錯誤',
          invalid_backup_code: '後備認證碼錯誤',
          too_many_authentication_attempts: '請稍後再試。'
        }
      },
      button: {
        sign_in: '登入',
        next: '下一步'
      },
      link: {
        forgot_password: '忘記密碼？',
        register: '建立帳戶',
        try_another_way: '使用其他方式',
        resend_verification_code: '再新發送認證碼'
      },
      text: {
        or: '或',
        password: '繼續用以下方式登入',
        code_sent: '認證碼已發送',
        forgot_password: '如果你忘記密碼，你可以使用24位回恢碼重新登入'
      }
    },
    reset_password: {
      title: {
        reset_password: '重設密碼',
        one_time_code_sent: '認證碼已發送',
        set_new_password: '設定新密碼',
        password_changed: '密碼已設定'
      },
      input: {
        label: {
          contact: '電郵或電話號碼',
          password: '密碼',
          confirm_password: '再次輸入密碼'
        },
        error: {
          blank_contact: '電郵或電話號碼不能為空',
          no_contact: '找不到相關電郵或電話號碼',
          reset_link_expired: '連結已過期'
        }
      },
      button: {
        send_reset_link: '發送重設連結',
        ok: '好',
        return_home: '返回登入頁面',
        reset_password: '重設密碼'
      },
      text: {
        email: '電郵',
        phone: '電話號碼',
        reset_link_sent: '重設密碼連結已發送至 {handle}',
        send_reset_password_instruction: '我們會把重設密碼的指示發送到',
        check_contact_instruction: '訊息將發送到{contact}，請確認。',
        reset_password_success: '你已成功重設密碼。請登入。',
        return_home: '返回登入頁面',
        error: {
          invalid_reset_password: '重設密碼連結無效',
          reach_limit: '抱歉！你已達到重設密碼的次數上限。如需協助，請與客戶服務部聯絡。'
        }
      }
    },
    profile: {
      list_item: {
        title: {
          profile: '個人檔案',
          contact: '聯絡方式'
        }
      },
      text: {
        and_more: '及更多'
      }
    },
    profile_edit: {
      title: '編輯個人檔案',
      input: {
        label: {
          name: '名稱'
        },
        error: {
          not_updated_username: '請輸入新的用戶名稱。'
        }
      },
      button: {
        save: '儲存',
        ok: '好'
      }
    },
    contacts: {
      // title: {contact} will be replaced by contact type, either phone or email
      title: ' 管理聯絡方式',
      button: {
        remove: '移除',
        // add_contact: {contact} will be replaced by contact type, either phone or email
        add_contact: '管理{contact}'
      },
      list_item: {
        text: {
          verified: '已認證',
          non_verified: '未認證'
        },
        link: {
          set_as_primary: '設定為預設聯絡方式',
          verify: '現在認證'
        }
      },
      text: {
        primary: '預設聯絡方式',
        email: '電郵',
        phone: '電話號碼'
      }
    },
    contact_create: {
      // title: {contact} will be replaced by contact type, either phone or email
      title: '管理{contact}',
      input: {
        label: {
          // contact: {contact} will be replaced by contact type, either phone or email
          contact: '{contact}'
        },
        error: {
          too_frequent: '請稍後再試',
          invalid_contact: '{contact}無效',
          duplicate_contact: '聯絡方式已有人使用'
        }
      },
      button: {
        verify: '認證',
        change: '更改'
      },
      text: {
        email: '電郵',
        phone: '電話號碼'
      }
    },
    contact_delete: {
      // title: {contact} will be replaced by contact type, either phone or email
      title: '移除{contact}',
      // description: {contact} will be replaced by contact type, either phone or email
      description: '你確認要移除{contact}嗎？',
      button: {
        remove: '移除',
        ok: '好'
      },
      text: {
        email: '電郵',
        phone: '電話號碼'
      }
    },
    contact_update_primary: {
      title: '設定為預設聯絡方式',
      // description: {contact} will be replaced by contact type, either phone or email
      description: '預設聯絡方式的作用為認證帳戶。確定要設定為{contact}嗎？',
      button: {
        ok: '好',
        update_primary_contact: '設定為預設聯絡方式'
      },
      text: {
        email: '電郵',
        phone: '電話號碼'
      }
    },
    settings_home: {
      title: '安全設定',
      list_item: {
        title: {
          change_password: '更改密碼',
          set_password: '設定密碼',
          password: '密碼',
          two_step_verification: '雙重認證',
          devices: '裝置',
          social_logins: '社群登入',
          admin: '管理版面',
          sign_out: '登出'
        },
        text: {
          off: '關閉',
          on: '開啟',
          manage_password: '管理密碼',
          change_password: '變更作為認證的密碼。',
          set_password: '設定作為認證的密碼。',
          two_step_verification: {
            add: '增設密碼加強安全性以及容許驗證設備',
            manage: '更多驗證設定加強安全性'
          },
          devices: '管理已登入裝置',
          social_logins: '管理社群登入',
          switch_to_admin: '進入管理版面'
        }
      }
    },
    set_password: {
      title: '設定密碼',
      button: {
        set_password: '設定密碼',
        ok: '好'
      },
      text: {
        password_set: '密碼已設定'
      }
    },
    change_password: {
      title: '變更密碼',
      input: {
        label: {
          old_password: '現用密碼',
          new_password: '新密碼',
          confirm_new_password: '確認新密碼'
        },
        error: {
          invalid_old_password: '現用密碼錯誤',
          requires_better_password_strength: '密碼強度不足',
          invalid_confirm_password: '密碼不符'
        }
      },
      button: {
        change_password: '變更密碼',
        ok: '好'
      },
      text: {
        password_updated: '密碼已變更'
      }
    },
    add_password: {
      title: '增加密碼',
      description: '增加密碼以加強安全性',
      input: {
        label: {
          password: '密碼',
          confirm_password: '確認密碼'
        }
      },
      button: {
        add_password: '增加密碼'
      }
    },
    manage_password: {
      title: '管理密碼',
      description: '更改或移除密碼',
      input: {
        label: {
          password: '密碼'
        }
      },
      button: {
        change_password: '更改密碼',
        remove_password: '移除密碼'
      }
    },
    modify_password: {
      input: {
        label: {
          password: '密碼',
          confirm_password: '確認密碼'
        }
      }
    },
    remove_password: {
      title: '移除密碼',
      description: '密碼移除後',
      button: {
        remove: '移除'
      }
    },
    mfa_list: {
      title: '管理雙重認證',
      description: '以一次認證碼及密碼登入',
      list_item: {
        title: {
          password: '密碼',
          authenticator_app: '認證應用程式'
        },
        text: {
          password: '管理密碼',
          security_key: '啟用生物及安全金鑰認證',
          authenticator_app: {
            create: '啟用認證應用程式接收六位數字認證碼',
            manage: '管理認證應用程式'
          },
          // added_on: {date} will be replaced by formatted date
          added_on: '新增日期：{date}',
          // last_generated: {date} will be replaced by formatted date
          last_generated: '上次使用日期：{date}'
        }
      }
    },
    manage_authenticator_app: {
      title: '移除認證應用程式',
      description: '現正使用認證應用程式作雙重認證',
      button: {
        remove: '移除認證應用程式'
      }
    },
    remove_authenticator_app: {
      title: '移除認證應用程式',
      description: '認證應用程式移除後你只能以一次認證碼及密碼登入',
      button: {
        remove: '移除認證應用程式'
      }
    },
    mfa_totp_create: {
      title: '設立認證應用程式',
      description: {
        scan_qrcode: '掃描QR碼',
        copy_key: '複製以下認證碼到認證應用程式。',
        // time_based_key: {0} will be replaced by bolded wording of Time based
        time_based_key: '請確定已設定為{0}。',
        time_based: '根據時間'
      },
      input: {
        label: {
          code: '認證碼'
        },
        error: {
          invalid_verification_code: '認證碼錯誤'
        }
      },
      button: {
        copy: '複製',
        copied: '已複製',
        next: '下一步'
      },
      link: {
        scan_qrcode: '掃描QR碼',
        set_up_manually: '手動設立'
      },
      text: {
        or: '或',
        // enter_authenticator_app_code: {0} will be replaced by bolded wording of how many digits of the code
        enter_authenticator_app_code: '輸入Authenticator App產生的{0}位認證碼'
      }
    },
    devices: {
      title: '管理裝置',
      list_item: {
        text: {
          // last_active: {date} will be replaced by formatted date
          last_active: '最後使用日期：{date}',
          this_device: '本裝置'
        }
      },
      button: {
        log_out: '登出',
        log_out_all_other_devices: '從所有其他裝置登出'
      }
    },
    device_delete: {
      title: '登出裝置',
      button: {
        log_out: '登出',
        ok: '好'
      },
      text: {
        log_out_the_following_device: '你確定要登出以下裝置嗎？',
        log_out_all_other_devices: '你確定要登出所有其他裝置嗎？',
        // last_active: {date} will be replaced by formatted date
        last_active: '最後使用日期：{date}'
      }
    },
    manage_social_logins: {
      title: '管理社群登入',
      description: {
        error: {
          used_contact_in_system: '此社交平台聯絡方式已被使用'
        }
      },
      list_item: {
        title: {
          google: 'Google',
          facebook: 'Facebook',
          twitter: 'Twitter',
          apple: 'Apple',
          matters: 'Matters'
        },
        text: {
          // last_used: {date} will be replaced by formatted date
          last_used: '最後使用日期：{date}',
          connect_now: '連結'
        }
      },
      button: {
        connect: '連結',
        disconnect: '取消連結'
      }
    },
    social_login_delete: {
      title: '取消連結社群登入？',
      button: {
        connect: '連結',
        disconnect: '取消連結',
        ok: '好'
      },
      text: {
        // disconnect_with_your_account: {platform} will be replaced by corresponding social platform
        // disconnect_with_your_account: 'Disconnect with your {platform} account?',
        disconnect_with_your_account: '你確定要取消連結以下社群登入嗎？',
        keep_one_social_login: '你應該最少保留一種社群登入方法',
        google: 'Google',
        facebook: 'Facebook',
        twitter: 'Twitter',
        apple: 'Apple',
        matters: 'Matters'
      }
    },
    error_page: {
      title: '嘗試使用{service}登入',
      text: {
        description: '你曾經使用{service}建立帳戶。請嘗試使用{service}登入',
        others: '其他方式',
        password: '密碼方式',
        google: 'Google',
        facebook: 'Facebook',
        twitter: 'Twitter',
        apple: 'Apple',
        matters: 'Matters'
      }
    },
    social_login_pane_list: {
      text: {
        register: '以{service}建立帳戶',
        sign_in: '以{service}登入',
        google: 'Google',
        facebook: 'Facebook',
        twitter: 'Twitter',
        apple: 'Apple',
        matters: 'Matters'
      }
    },
    // Section below refers to commonly used translation value for the whole application.
    general: {
      // blank: Javascript blank character, shall be converted into blank character in HTML to show the element without content. Mainly used for consistent spacing.
      blank: '\xa0',
      separator: '或'
    },
    description: {
      work_in_progress: '準備中',

      // n_digits: {0} will be replaced by how many digits for the code
      n_digits: '{0}位',
      verifying: '認證中',
      completed: '完成',
      // protected_by: Formatted footer
      protected_by: '{start} {logo}{name}'
    },
    button: {
      ok: '好',
      confirm: '確定'
    },
    error: {
      unknown: '未知錯誤',
      accept_privacy_policy: '本欄必須確認',
      blank: '本欄不能為空'
    }
  }
}
