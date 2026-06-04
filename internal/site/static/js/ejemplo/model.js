// Model de un usuario del JSON público de demo.
//
// Subclasear Model vale la pena para NORMALIZAR la forma cruda de la API (acá,
// aplanar el objeto anidado `company`) y exponer accesores con sentido de
// dominio, en vez de cavar en `.get('company').name` por toda la vista.
import { Model } from 'lumen';

export class User extends Model {
  /** Nombre de la empresa, aplanado del objeto anidado que devuelve la API. */
  companyName() {
    return this.get('company')?.name ?? '—';
  }

  /** Iniciales para el avatar (sin pedir imágenes externas): 1ª letra de las dos primeras palabras. */
  initials() {
    return (this.get('name') || '?')
      .split(/\s+/)
      .slice(0, 2)
      .map((word) => word[0])
      .join('')
      .toUpperCase();
  }

  /** Matiz (0-359) estable derivado del nombre, para colorear el avatar de forma consistente. */
  hue() {
    const name = this.get('name') || '';
    let hash = 0;
    for (let i = 0; i < name.length; i++) hash = (hash * 31 + name.charCodeAt(i)) % 360;
    return hash;
  }
}
