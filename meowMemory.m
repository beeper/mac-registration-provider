#include "meowMemory.h"

NSAutoreleasePool* meowMakePool() {
	return [[NSAutoreleasePool alloc] init];
}
void meowReleasePool(NSAutoreleasePool* pool) {
	[pool drain];
}