#include "IMDAppleServices.h"
#include <Foundation/Foundation.h>
#include <dlfcn.h>

void *HANDLE;
void *BASE;

// Loads the IMDAppleServices framework and calculates the base address
void NACLoad() {
  if (!HANDLE) {
    NSLog(@"Loading framework %s...", IMD_PATH);
    HANDLE = dlopen(IMD_PATH, RTLD_LAZY);
    if (!HANDLE) {
      NSLog(@"dlopen failed: %s", dlerror());
      exit(-1);
    }
    NSLog(@"Successfully loaded %s", IMD_PATH);
  }

  if (!BASE) {
    NSLog(@"Using reference symbol %s at %p to calculate base address...",
          IMD_REF_SYM, (void *)IMD_REF_ADDR);
    void *ref = dlsym(HANDLE, IMD_REF_SYM);
    if (!ref) {
      NSLog(@"dlsym failed: %s", dlerror());
      exit(-1);
    }
    BASE = ref - IMD_REF_ADDR;
    NSLog(@"Calculated base address: %p", BASE);
  }
}

int NACInit(const void *cert_bytes, int cert_len, void **out_validation_ctx,
            void **out_request_bytes, int *out_request_len) {
  if (!HANDLE || !BASE) {
    NACLoad();
  }

  int (*nac_init)(void *, int, void *, void *, void *) =
      BASE + IMD_NACINIT_ADDR;
  return nac_init((void *)cert_bytes, cert_len, out_validation_ctx,
                  out_request_bytes, out_request_len);
}

int NACKeyEstablishment(void *validation_ctx, void *response_bytes, int response_len) {
  if (!HANDLE || !BASE) {
    NACLoad();
  }

  int (*nac_submit)(void *, void *, int) = BASE + IMD_NACSUBMIT_ADDR;
  return nac_submit(validation_ctx, response_bytes, response_len);
}

// No idea what unk_bytes is for, you can pass NULL
int NACSign(void *validation_ctx, void *unk_bytes, int unk_len,
                void **validation_data, int *validation_data_len) {
  if (!HANDLE || !BASE) {
    NACLoad();
  }

  int (*nac_generate)(void *, void *, int, void *, int *) =
      BASE + IMD_NACGENERATE_ADDR;
  return nac_generate(validation_ctx, unk_bytes, unk_len, validation_data,
                      validation_data_len);
}
