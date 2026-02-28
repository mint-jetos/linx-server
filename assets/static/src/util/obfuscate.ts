export const obfuscationKey = "linx-evasion-key";

export const xorTransform = (data: Uint8Array, offset: number = 0): Uint8Array => {
  const result = new Uint8Array(data.length);
  for (let i = 0; i < data.length; i++) {
    result[i] = data[i]! ^ obfuscationKey.charCodeAt((i + offset) % obfuscationKey.length);
  }
  return result;
};

export const obfuscate = (text: string): string => {
  const encoder = new TextEncoder();
  const bytes = encoder.encode(text);
  const transformed = xorTransform(bytes, 0);
  let binary = "";
  for (let i = 0; i < transformed.length; i++) {
    binary += String.fromCharCode(transformed[i]!);
  }
  return btoa(binary);
};

export const deobfuscate = (encoded: string): string => {
  const binaryString = atob(encoded);
  const bytes = new Uint8Array(binaryString.length);
  for (let i = 0; i < binaryString.length; i++) {
    bytes[i] = binaryString.charCodeAt(i);
  }
  const originalBytes = xorTransform(bytes, 0);
  const decoder = new TextDecoder();
  return decoder.decode(originalBytes);
};
