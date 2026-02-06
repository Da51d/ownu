export interface User {
  id: string
  username: string
}

export interface AuthResponse {
  token: string
  user: User
}

export interface Account {
  id: string
  name: string
  type: string
  institution: string
  createdAt: string
  updatedAt: string
}

export interface Transaction {
  id: string
  accountId: string
  amount: number
  description: string
  merchant: string
  date: string
  categoryId?: string
  createdAt: string
}

export interface Category {
  id: string
  name: string
  parentId?: string
  isSystem: boolean
}

export interface RegisterBeginResponse {
  options: PublicKeyCredentialCreationOptions
  recovery_phrase: string
  session_id: string
}

export interface LoginBeginResponse {
  options: PublicKeyCredentialRequestOptions
  session_id: string
}
