// NB: string types intentionaly changed to bytes

req_pq_multi#be7e8ef1 nonce:int128 = ReqPQMulti;

resPQ#05162463 nonce:int128 server_nonce:int128 pq:bytes server_public_key_fingerprints:vector<long> = ResPQ;

p_q_inner_data_const#83c95aec pq:bytes p:bytes q:bytes nonce:int128 server_nonce:int128 new_nonce:int256 = P_Q_inner_data;

p_q_inner_data_temp#3c6a84d4 pq:bytes p:bytes q:bytes nonce:int128 server_nonce:int128 new_nonce:int256 expires_in:int = P_Q_inner_data;

server_DH_inner_data#b5890dba nonce:int128 server_nonce:int128 g:int dh_prime:bytes g_a:bytes server_time:int = Server_DH_inner_data;

client_DH_inner_data#6643b654 nonce:int128 server_nonce:int128 retry_id:long g_b:bytes = Client_DH_Inner_Data;
