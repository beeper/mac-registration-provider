#include <Foundation/Foundation.h>

int nacInitProxy(void *addr, const void *cert_bytes, int cert_len, void **out_validation_ctx, void **out_request_bytes, int *out_request_len);
int nacKeyEstablishmentProxy(void *addr, void *validation_ctx, void *response_bytes, int response_len);
int nacSignProxy(void *addr, void *validation_ctx, void *unk_bytes, int unk_len, void **validation_data, int *validation_data_len);

NSAutoreleasePool* meowMakePool();
void meowReleasePool(NSAutoreleasePool* pool);
