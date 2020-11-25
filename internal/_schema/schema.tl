// NB: string types intentionaly changed to bytes

req_pq_multi#be7e8ef1 nonce:int128 = ReqPQMulti;

resPQ#05162463 nonce:int128 server_nonce:int128 pq:bytes server_public_key_fingerprints:vector<long> = ResPQ;

p_q_inner_data_const#83c95aec pq:bytes p:bytes q:bytes nonce:int128 server_nonce:int128 new_nonce:int256 = P_Q_inner_data;

p_q_inner_data_temp#3c6a84d4 pq:bytes p:bytes q:bytes nonce:int128 server_nonce:int128 new_nonce:int256 expires_in:int = P_Q_inner_data;

server_DH_inner_data#b5890dba nonce:int128 server_nonce:int128 g:int dh_prime:bytes g_a:bytes server_time:int = Server_DH_inner_data;

client_DH_inner_data#6643b654 nonce:int128 server_nonce:int128 retry_id:long g_b:bytes = Client_DH_Inner_Data;

req_DH_params#d712e4be nonce:int128 server_nonce:int128 p:bytes q:bytes public_key_fingerprint:long encrypted_data:bytes = Req_DH_params;

server_DH_params_fail#79cb045d nonce:int128 server_nonce:int128 new_nonce_hash:int128 = Server_DH_Params;

server_DH_params_ok#d0e8075c nonce:int128 server_nonce:int128 encrypted_answer:bytes = Server_DH_Params;

set_client_DH_params#f5045f1f nonce:int128 server_nonce:int128 encrypted_data:bytes = Set_client_DH_params;

dh_gen_ok#3bcbf734 nonce:int128 server_nonce:int128 new_nonce_hash1:int128 = Set_client_DH_params_answer;
dh_gen_retry#46dc1fb9 nonce:int128 server_nonce:int128 new_nonce_hash2:int128 = Set_client_DH_params_answer;
dh_gen_fail#a69dae02 nonce:int128 server_nonce:int128 new_nonce_hash3:int128 = Set_client_DH_params_answer;
