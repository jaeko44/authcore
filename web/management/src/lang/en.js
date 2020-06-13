export default {
  en: {
    model: {
      user: {
        password: 'Password',
        confirm_password: 'Confirm password',
        name: 'Name',
        username: 'Username',
        email: 'Email',
        verified: 'Verified',
        phone: 'Phone',
        email_or_phone: 'Email or phone',
        language: 'Language',
        created_at: 'Created date',
        last_seen_at: 'Last login date',
        no_last_seen_at: 'No login yet',
        locked_user: 'Locked user',
        locked: 'Locked',
        profile: 'Profile'
      },
      event: {
        action: 'Event',
        status: 'Status',
        ip: 'IP',
        device: 'Device',
        created_at: 'When',
        result: {
          unknown: 'Unknown',
          success: 'Success',
          fail: 'Fail'
        }
      },
      second_factor: {
        sms_otp: 'SMS',
        totp: 'Authenticator App',
        backup_code: 'Backup code',
        value: 'Description',
        created_at_with_date: 'Since {date}',
        last_used_at: 'Last used'
      },
      oauth_factor: {
        social_media: 'Social media',
        platform_user_id: 'Social platform User ID',
        created_at: 'Associated at',
        created_at_with_date: 'Since {date}',
        last_used_at: 'Last used',
        service: {
          google: 'Google',
          facebook: 'Facebook',
          twitter: 'Twitter',
          apple: 'Apple',
          matters: 'Matters'
        }
      },
      device: {
        user_agent: 'User agent',
        last_seen_at: 'Last active',
        last_seen_at_with_date: 'Last active: {date}'
      },
      role: {
        name: 'Role name',
        description: 'Description',
        admin: 'Admin',
        editor: 'Editor'
      },
      template: {
        updated_at: 'Last edit time',
        default: 'Default',
        description: 'Description',
        subject: 'Subject',
        html_template: 'HTML template',
        text_template: 'Text template',
        type: {
          authentication_sms: {
            name: 'Authentication',
            description: 'To authenticate user via SMS.'
          },
          verification_sms: {
            name: 'Verification',
            description: 'To confirm the account is setup.'
          },
          reset_password_authentication_sms: {
            name: 'Reset password',
            description: 'To assist user to reset password.'
          },
          verification_mail: {
            name: 'Verification',
            description: 'To confirm the account is setup.'
          },
          reset_password_authentication_mail: {
            name: 'Reset password',
            description: 'To assist user to reset password.'
          }
        }
      }
    },
    error: {
      unknown: 'Unknown',
      role_assignment_exists_for_same_account: 'Role already assigned to the user.',
      role_unassignment_for_same_account: 'Cannot unassign role for the same account.',
      invalid_user_metadata_syntax: 'User metadata syntax error.',
      invalid_app_metadata_syntax: 'App metadata syntax error.',
      invalid_username: 'Invalid username',
      duplicate_username: 'Duplicate username',
      invalid_email: 'Invalid email',
      invalid_phone: 'Invalid phone',
      invalid_contact: 'Invalid contact',
      duplicate_contact: 'Duplicate contact',
      no_permission: 'Current user does not have permission to access this page.'
    },
    common: {
      blank: '\xa0',
      product_name: 'Authcore Management',
      on: 'ON',
      off: 'OFF',
      total_items: '{count} items',
      total_users: '{count} user | {count} users',
      languages: {
        en: 'English',
        'zh-HK': '繁體中文（香港）'
      },
      '2fa': '2-step verification',
      '2fa_short': '2FA',
      social_login: 'Social login',
      work_in_progress: 'Work in progress'
    },
    navigation: {
      title: {
        users: 'Users',
        settings: 'Settings',
        email_settings: 'Email',
        sms_settings: 'SMS',
        user_portal: 'Manage your account'
      }
    },
    data_table: {
      text: {
        first_page: 'First page'
      }
    },
    user_list: {
      meta: {
        title: 'User List'
      },
      title: 'Users',
      message: {
        user_created_successfully: 'User created successfully.',
        user_deleted_successfully: 'User deleted successfully.'
      },
      text: {
        all: 'All',
        search_for_user: 'Search for user'
      },
      button: {
        create_user: 'Create user'
      }
    },
    user_details: {
      meta: {
        title: 'User Details'
      },
      title: 'User details',
      sub_title: {
        profile: 'Profile',
        events: 'Events',
        security: 'Security',
        devices: 'Devices',
        roles: 'Roles',
        metadata: 'Metadata'
      },
      actions: {
        lock_user: 'Lock user',
        unlock_user: 'Unlock user',
        delete_user: 'Delete user'
      },
      button: {
        actions: 'Actions'
      }
    },
    user_details_profile: {
      meta: {
        title: 'Profile - User Details'
      },
      title: 'Profile',
      description: 'Edit user profile',
      message: {
        update_user_successfully: 'Update user successfully.'
      },
      button: {
        save: 'Save'
      }
    },
    user_details_events: {
      title: 'Events',
      description: 'Max. logs of the last month.',
      table: {
        header: {
          action: 'Event',
          status: 'Status',
          ip: 'IP',
          device: 'Device',
          when: 'When'
        }
      },
      meta: {
        title: 'Logs - User Details'
      }
    },
    user_details_security: {
      meta: {
        title: 'Security - User Details'
      },
      message: {
        change_password_successfully: 'Change password successfully.'
      },
      text: {
        unlink_unavailable: 'Cannot unlink this account',
        no_social_account_linked: 'No social account linked'
      },
      button: {
        unlink: 'Unlink',
        set_password: 'Set password',
        change_password: 'Change password'
      }
    },
    user_details_devices: {
      meta: {
        title: 'Devices - User Details'
      },
      title: 'Devices',
      description: 'Logged in devices',
      text: {
        log_out_device: 'Log out device'
      },
      button: {
        log_out: 'Log out',
        log_out_all: 'Log out all devices'
      }
    },
    user_details_roles: {
      meta: {
        title: 'Roles - User Details'
      },
      title: 'Roles',
      description: 'All roles assigned to user.',
      text: {
        assign_role: 'Assign role',
        no_roles_assigned: 'User can access basic features',
        admin: 'Manage the admin panel',
        editor: ''
      },
      button: {
        unassign: 'Unassign'
      }
    },
    user_details_metadata: {
      meta: {
        title: 'Metadata - User Details'
      },
      title: {
        app_metadata: 'App metadata',
        user_metadata: 'User metadata'
      },
      description: {
        user_metadata: 'Data that the user has read/write access to (e.g. color_preference, blog_url, etc.)',
        app_metadata: 'Data that the user has read-only access to (e.g. roles, permissions, vip, etc.)'
      },
      message: {
        update_metadata: 'Update metadata success.'
      },
      button: {
        save: 'Save'
      }
    },
    user_create: {
      meta: {
        title: 'Create new user'
      },
      title: 'Create user',
      description: 'Profile',
      text: {
        back_title: 'Users',
        generate_password: 'Generate password'
      },
      button: {
        create_user: 'Create user',
        generate: 'Generate',
        copy: 'Copy'
      }
    },
    settings: {
      meta: {
        title: 'Settings'
      }
    },
    email_template_list: {
      meta: {
        title: 'Email Template Settings'
      },
      title: 'Email settings',
      sub_title: 'Template',
      description: 'Customise the account verification, reset password and security alert emails to fit your branding.',
      text: {
        email_templates: 'Email templates'
      },
      button: {
        edit: 'Edit'
      }
    },
    email_template_edit: {
      title: 'Edit {name} email',
      message: {
        success_create: 'Template updated successfully',
        success_reset: 'Template reset successfully'
      },
      button: {
        save: 'Save',
        reset: 'Reset'
      },
      modal: {
        title: 'Reset',
        description: 'Are you sure to reset template?',
        button: {
          confirm: 'Confirm'
        }
      }
    },
    sms_template_list: {
      meta: {
        title: 'SMS Template Settings'
      },
      title: 'SMS settings',
      sub_title: 'Template',
      description: 'Customise the account verification, reset password and security alert SMSs to fit your branding.',
      text: {
        sms_templates: 'SMS templates'
      },
      button: {
        edit: 'Edit'
      }
    },
    sms_template_edit: {
      title: 'Edit {name} SMS',
      message: {
        success_create: 'Template updated successfully',
        success_reset: 'Template reset successfully'
      },
      button: {
        save: 'Save',
        reset: 'Reset'
      },
      modal: {
        title: 'Reset',
        description: 'Are you sure to reset template?',
        button: {
          confirm: 'Confirm'
        }
      }
    },
    change_password_modal_pane: {
      title: {
        change_password: 'Change password',
        set_password: 'Set password'
      },
      description: {
        change_password: 'Change the password for {username}',
        set_password: 'Set the password for {username}'
      },
      button: {
        change_password: 'Change password',
        set_password: 'Set password'
      }
    },
    lock_user_modal_pane: {
      title: {
        lock_user: 'Lock user',
        unlock_user: 'Unlock user'
      },
      description: {
        lock_user: 'Confirm to lock user?',
        unlock_user: 'Confirm to unlock user?'
      },
      button: {
        lock_user: 'Lock user',
        unlock_user: 'Unlock user'
      }
    },
    logout_device_modal_pane: {
      title: 'Log out device',
      description: 'The following device will be logged out:',
      button: {
        log_out: 'Log out'
      }
    },
    unlink_oauth_factor_modal_pane: {
      title: 'Unlink {service} OAuth factor',
      description: 'Confirm to unlink {service} OAuth factor?',
      button: {
        unlink_oauth_factor: 'Unlink {service} OAuth factor'
      }
    },
    unlink_second_factor_modal_pane: {
      title: 'Unlink {type} second factor',
      description: 'Confirm to unlink {type} second factor?',
      button: {
        unlink_second_factor: 'Unlink {type} second factor'
      }
    },
    delete_user_modal_pane: {
      title: 'Delete user',
      description: 'Confirm to delete user?',
      button: {
        delete_user: 'Delete user'
      }
    }
  }
}
