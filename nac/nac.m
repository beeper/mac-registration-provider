#include "nac.h"

int nacInitProxy(void *addr, const void *cert_bytes, int cert_len, void **out_validation_ctx, void **out_request_bytes, int *out_request_len) {
  int (*nac_init)(void *, int, void *, void *, void *) = addr;
  return nac_init((void *)cert_bytes, cert_len, out_validation_ctx, out_request_bytes, out_request_len);
}

int nacKeyEstablishmentProxy(void *addr, void *validation_ctx, void *response_bytes, int response_len) {
  int (*nac_key_establishment)(void *, void *, int) = addr;
  return nac_key_establishment(validation_ctx, response_bytes, response_len);
}

// No idea what unk_bytes is for, you can pass NULL
int nacSignProxy(void *addr, void *validation_ctx, void *unk_bytes, int unk_len, void **validation_data, int *validation_data_len) {
  int (*nac_sign)(void *, void *, int, void *, int *) = addr;
  return nac_sign(validation_ctx, unk_bytes, unk_len, validation_data, validation_data_len);
}

NSAutoreleasePool* meowMakePool() {
	return [[NSAutoreleasePool alloc] init];
}
void meowReleasePool(NSAutoreleasePool* pool) {
	[pool drain];
}
