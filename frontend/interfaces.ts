export interface ImportRecord {
    id: number;
    useDate: string;
    name: string;
    price: number;
    registDate: string;
    importJudgeDate: string;
    importDate: string;
}

export interface Rule {
    id: number;
    fieldName: string;
    value: number;
    exactMatch: number;
    categoryId: number;
}
